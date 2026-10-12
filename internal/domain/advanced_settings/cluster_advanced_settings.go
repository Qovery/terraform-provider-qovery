package advanced_settings

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"
)

type ClusterAdvancedSettingsService struct {
	apiConfig *qovery.Configuration

	// defaultKeysCache caches the set of valid cluster advanced setting keys. The default
	// set is static for a provider run. It is held behind a pointer so that value-receiver
	// method copies share the same cache.
	defaultKeysCache *clusterDefaultKeysCache
}

// clusterDefaultKeysCache holds the lazily-fetched set of valid cluster advanced setting keys.
type clusterDefaultKeysCache struct {
	mu   sync.Mutex
	keys map[string]struct{}
}

// apiHTTPClient returns the http client of the Qovery API configuration, so the advanced settings
// calls get the same retry policy as the generated client.
func apiHTTPClient(apiConfig *qovery.Configuration) *http.Client {
	if apiConfig.HTTPClient != nil {
		return apiConfig.HTTPClient
	}
	return http.DefaultClient
}

func NewClusterAdvancedSettingsService(apiConfig *qovery.Configuration) *ClusterAdvancedSettingsService {
	return &ClusterAdvancedSettingsService{
		apiConfig:        apiConfig,
		defaultKeysCache: &clusterDefaultKeysCache{},
	}
}

// fetchDefaultClusterAdvancedSettings fetches and parses the default cluster advanced
// settings, whose keys form the set of valid cluster advanced setting keys.
func (c ClusterAdvancedSettingsService) fetchDefaultClusterAdvancedSettings(ctx context.Context) (map[string]any, error) {
	httpClient := apiHTTPClient(c.apiConfig)
	apiToken := c.apiConfig.DefaultHeader["Authorization"]
	host := c.apiConfig.Servers[0].URL

	defaultAdvancedSettingsUrl := host + "/defaultClusterAdvancedSettings"
	getDefaultAdvancedSettingsRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, defaultAdvancedSettingsUrl, nil)
	if err != nil {
		return nil, err
	}
	getDefaultAdvancedSettingsRequest.Header.Set("Authorization", apiToken)
	getDefaultAdvancedSettingsRequest.Header.Set("Content-Type", "application/json")
	getDefaultAdvancedSettingsRequest.Header.Set("User-Agent", c.apiConfig.UserAgent)

	respGetDefaultAdvancedSettings, err := httpClient.Do(getDefaultAdvancedSettingsRequest)
	if err != nil {
		return nil, err
	}
	defer respGetDefaultAdvancedSettings.Body.Close()

	if respGetDefaultAdvancedSettings.StatusCode >= 400 {
		return nil, errors.New("Cannot get default cluster advanced settings :" + respGetDefaultAdvancedSettings.Status)
	}

	clusterDefaultAdvancedSettings, err := io.ReadAll(respGetDefaultAdvancedSettings.Body)
	if err != nil {
		return nil, err
	}

	defaultAdvancedSettingsHashMap := make(map[string]any)
	if err := json.Unmarshal(clusterDefaultAdvancedSettings, &defaultAdvancedSettingsHashMap); err != nil {
		return nil, err
	}

	return defaultAdvancedSettingsHashMap, nil
}

// defaultSettingKeys returns the set of valid cluster advanced setting keys, fetched once
// and cached for the lifetime of the service.
func (c ClusterAdvancedSettingsService) defaultSettingKeys(ctx context.Context) (map[string]struct{}, error) {
	cache := c.defaultKeysCache

	cache.mu.Lock()
	if cache.keys != nil {
		cached := cache.keys
		cache.mu.Unlock()
		return cached, nil
	}
	cache.mu.Unlock()

	defaults, err := c.fetchDefaultClusterAdvancedSettings(ctx)
	if err != nil {
		return nil, err
	}

	keys := make(map[string]struct{}, len(defaults))
	for k := range defaults {
		keys[k] = struct{}{}
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.keys == nil {
		cache.keys = keys
	}
	return cache.keys, nil
}

// UnknownSettingKeys returns the advanced setting keys present in advancedSettingsJson that
// are not valid cluster advanced settings (absent from the default cluster settings). It
// returns nil for an empty input.
func (c ClusterAdvancedSettingsService) UnknownSettingKeys(ctx context.Context, advancedSettingsJson string) ([]string, error) {
	if advancedSettingsJson == "" || advancedSettingsJson == "{}" {
		return nil, nil
	}

	validKeys, err := c.defaultSettingKeys(ctx)
	if err != nil {
		return nil, err
	}

	provided := make(map[string]any)
	if err := json.Unmarshal([]byte(advancedSettingsJson), &provided); err != nil {
		return nil, err
	}

	return computeUnknownKeys(validKeys, provided), nil
}

// ReadClusterAdvancedSettings returns only overridden advanced settings.
func (c ClusterAdvancedSettingsService) ReadClusterAdvancedSettings(
	ctx context.Context,
	organizationId string,
	clusterId string,
	advancedSettingsJsonFromState string,
	isTriggeredFromImport bool,
) (*string, error) {
	httpClient := apiHTTPClient(c.apiConfig)
	apiToken := c.apiConfig.DefaultHeader["Authorization"]
	host := c.apiConfig.Servers[0].URL

	var clusterAdvancedSettingsState string
	if advancedSettingsJsonFromState == "" {
		clusterAdvancedSettingsState = "{}"
	} else {
		clusterAdvancedSettingsState = advancedSettingsJsonFromState
	}
	//
	// Get default cluster advanced settings
	defaultAdvancedSettingsHashMap, err := c.fetchDefaultClusterAdvancedSettings(ctx)
	if err != nil {
		return nil, err
	}

	//
	// Get cluster advanced settings
	urlAdvancedSettings := host + "/organization/" + organizationId + "/cluster/" + clusterId + "/advancedSettings"
	getRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, urlAdvancedSettings, nil)
	if err != nil {
		return nil, err
	}
	getRequest.Header.Set("Authorization", apiToken)
	getRequest.Header.Set("Content-Type", "application/json")
	getRequest.Header.Set("User-Agent", c.apiConfig.UserAgent)

	respGetAdvancedSettings, err := httpClient.Do(getRequest)
	if err != nil {
		return nil, err
	}
	defer respGetAdvancedSettings.Body.Close()

	if respGetAdvancedSettings.StatusCode >= 400 {
		return nil, errors.New("Cannot get cluster advanced settings :" + respGetAdvancedSettings.Status)
	}

	clusterAdvancedSettings, err := io.ReadAll(respGetAdvancedSettings.Body)
	if err != nil {
		return nil, err
	}

	advancedSettingsStringJson := string(clusterAdvancedSettings)

	//
	// Compute the Diff
	advancedSettingsFromStateHashMap := make(map[string]any)
	if err := json.Unmarshal([]byte(clusterAdvancedSettingsState), &advancedSettingsFromStateHashMap); err != nil {
		return nil, err
	}

	currentAdvancedSettingsHashMap := make(map[string]any)
	if err := json.Unmarshal([]byte(advancedSettingsStringJson), &currentAdvancedSettingsHashMap); err != nil {
		return nil, err
	}

	overriddenAdvancedSettings := computeOverriddenSettings(
		currentAdvancedSettingsHashMap,
		defaultAdvancedSettingsHashMap,
		advancedSettingsFromStateHashMap,
		isTriggeredFromImport,
	)

	//
	// Transform to JSON
	overriddenAdvancedSettingsJSON, err := json.Marshal(overriddenAdvancedSettings)
	if err != nil {
		return nil, errors.New("Cannot parse overridden cluster advanced settings")
	}

	s := string(overriddenAdvancedSettingsJSON)
	return &s, nil
}

// UpdateClusterAdvancedSettings updates advanced settings by computing the whole HTTP body.
func (c ClusterAdvancedSettingsService) UpdateClusterAdvancedSettings(ctx context.Context, organizationId string, clusterId string, advancedSettingsJsonParam string) error {
	httpClient := apiHTTPClient(c.apiConfig)
	apiToken := c.apiConfig.DefaultHeader["Authorization"]
	host := c.apiConfig.Servers[0].URL

	var advancedSettingsJson string
	if advancedSettingsJsonParam == "" {
		advancedSettingsJson = "{}"
	} else {
		advancedSettingsJson = advancedSettingsJsonParam
	}

	//
	// Get cluster advanced settings
	urlAdvancedSettings := host + "/organization/" + organizationId + "/cluster/" + clusterId + "/advancedSettings"
	overriddenAdvancedSettingsHashMap := make(map[string]any)
	if err := json.Unmarshal([]byte(advancedSettingsJson), &overriddenAdvancedSettingsHashMap); err != nil {
		return err
	}

	getRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, urlAdvancedSettings, nil)
	if err != nil {
		return err
	}
	getRequest.Header.Set("Authorization", apiToken)
	getRequest.Header.Set("Content-Type", "application/json")
	getRequest.Header.Set("User-Agent", c.apiConfig.UserAgent)

	respGetAdvancedSettings, err := httpClient.Do(getRequest)
	if err != nil {
		return err
	}
	defer respGetAdvancedSettings.Body.Close()

	if respGetAdvancedSettings.StatusCode >= 400 {
		return errors.New("Cannot get cluster advanced settings :" + respGetAdvancedSettings.Status)
	}
	clusterAdvancedSettings, err := io.ReadAll(respGetAdvancedSettings.Body)
	if err != nil {
		return err
	}

	advancedSettingsStringJson := string(clusterAdvancedSettings)

	//
	// Compute final http body to send to satisfy PUT endpoint
	currentAdvancedSettingsHashMap := make(map[string]any)
	if err := json.Unmarshal([]byte(advancedSettingsStringJson), &currentAdvancedSettingsHashMap); err != nil {
		return err
	}

	for k, v := range currentAdvancedSettingsHashMap {
		_, exists := overriddenAdvancedSettingsHashMap[k]
		if !exists {
			overriddenAdvancedSettingsHashMap[k] = v
		}
	}

	overriddenAdvancedSettingsJSON, err := json.Marshal(overriddenAdvancedSettingsHashMap)
	if err != nil {
		return err
	}

	//
	// Update advanced settings
	putRequest, err := http.NewRequestWithContext(ctx, http.MethodPut, urlAdvancedSettings, bytes.NewBuffer(overriddenAdvancedSettingsJSON))
	if err != nil {
		return err
	}
	putRequest.Header.Set("Authorization", apiToken)
	putRequest.Header.Set("Content-Type", "application/json")
	putRequest.Header.Set("Accept", "application/json")
	putRequest.Header.Set("User-Agent", c.apiConfig.UserAgent)

	respPostAdvancedSettings, err := httpClient.Do(putRequest)
	if err != nil {
		return err
	}

	defer respPostAdvancedSettings.Body.Close()

	if respPostAdvancedSettings.StatusCode >= 400 {
		body, _ := io.ReadAll(respPostAdvancedSettings.Body)
		return errors.New("Cannot update cluster advanced settings :" + string(body))
	}

	return nil
}
