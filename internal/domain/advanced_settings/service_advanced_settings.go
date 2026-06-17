package advanced_settings

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"sync"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain"
)

type ServiceAdvancedSettingsService struct {
	apiConfig *qovery.Configuration

	// defaultKeysCache caches the set of valid advanced setting keys per service type.
	// The default set is static for a provider run. It is a reference-type map guarded by
	// a pointer mutex so that value-receiver method copies share the same cache.
	defaultKeysCache map[int]map[string]struct{}
	cacheMu          *sync.Mutex
}

func NewServiceAdvancedSettingsService(apiConfig *qovery.Configuration) *ServiceAdvancedSettingsService {
	return &ServiceAdvancedSettingsService{
		apiConfig:        apiConfig,
		defaultKeysCache: make(map[int]map[string]struct{}),
		cacheMu:          &sync.Mutex{},
	}
}

// Compute the URL to GET or PUT advanced settings for any service type
func (c ServiceAdvancedSettingsService) computeServiceAdvancedSettingsUrl(serviceType int, serviceId string) (*string, error) {
	host := c.apiConfig.Servers[0].URL
	var urlAdvancedSettings string

	switch serviceType {
	case domain.APPLICATION:
		urlAdvancedSettings = host + "/application/" + serviceId + "/advancedSettings"
	case domain.CONTAINER:
		urlAdvancedSettings = host + "/container/" + serviceId + "/advancedSettings"
	case domain.JOB:
		urlAdvancedSettings = host + "/job/" + serviceId + "/advancedSettings"
	case domain.HELM:
		urlAdvancedSettings = host + "/helm/" + serviceId + "/advancedSettings"
	case domain.TERRAFORM:
		urlAdvancedSettings = host + "/terraform/" + serviceId + "/advancedSettings"
	default:
		return nil, errors.New("serviceType must be one of APPLICATION / CONTAINER / JOB / HELM / TERRAFORM")
	}

	return &urlAdvancedSettings, nil
}

// Compute the URL to GET default advanced settings for any service type
func (c ServiceAdvancedSettingsService) computeDefaultServiceAdvancedSettingsUrl(serviceType int) (*string, error) {
	host := c.apiConfig.Servers[0].URL
	var urlAdvancedSettings string

	switch serviceType {
	case domain.APPLICATION:
		urlAdvancedSettings = host + "/defaultApplicationAdvancedSettings"
	case domain.CONTAINER:
		urlAdvancedSettings = host + "/defaultContainerAdvancedSettings"
	case domain.JOB:
		urlAdvancedSettings = host + "/defaultJobAdvancedSettings"
	case domain.HELM:
		urlAdvancedSettings = host + "/defaultHelmAdvancedSettings"
	case domain.TERRAFORM:
		urlAdvancedSettings = host + "/defaultTerraformAdvancedSettings"
	default:
		return nil, errors.New("serviceType must be one of APPLICATION / CONTAINER / JOB / HELM / TERRAFORM")
	}

	return &urlAdvancedSettings, nil
}

// ReadServiceAdvancedSettings Get only overridden advanced settings
func (c ServiceAdvancedSettingsService) ReadServiceAdvancedSettings(serviceType int, serviceId string, advancedSettingsJsonFromState string, isTriggeredFromImport bool) (*string, error) {
	httpClient := &http.Client{}
	apiToken := c.apiConfig.DefaultHeader["Authorization"]

	var serviceAdvancedSettingsState string
	if advancedSettingsJsonFromState == "" {
		serviceAdvancedSettingsState = "{}"
	} else {
		serviceAdvancedSettingsState = advancedSettingsJsonFromState
	}

	//
	// Get service advanced settings
	urlAdvancedSettings, err := c.computeServiceAdvancedSettingsUrl(serviceType, serviceId)
	if err != nil {
		return nil, err
	}
	getRequest, err := http.NewRequest("GET", *urlAdvancedSettings, nil)
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
		return nil, errors.New("Cannot get advanced settings :" + respGetAdvancedSettings.Status)
	}

	serviceAdvancedSettings, err := io.ReadAll(respGetAdvancedSettings.Body)
	if err != nil {
		return nil, err
	}

	advancedSettingsStringJson := string(serviceAdvancedSettings)

	//
	// Compute the Diff
	advancedSettingsFromStateHashMap := make(map[string]any)
	if err := json.Unmarshal([]byte(serviceAdvancedSettingsState), &advancedSettingsFromStateHashMap); err != nil {
		return nil, err
	}

	currentAdvancedSettingsHashMap := make(map[string]any)
	if err := json.Unmarshal([]byte(advancedSettingsStringJson), &currentAdvancedSettingsHashMap); err != nil {
		return nil, err
	}

	defaultAdvancedSettingsHashMap, err := c.fetchDefaultAdvancedSettings(serviceType)
	if err != nil {
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
		return nil, errors.New("Cannot parse overridden advanced settings")
	}

	s := string(overriddenAdvancedSettingsJSON)
	return &s, nil
}

// UpdateServiceAdvancedSettings Update advanced settings by computing the whole http body
func (c ServiceAdvancedSettingsService) UpdateServiceAdvancedSettings(serviceType int, serviceId string, advancedSettingsJsonFromPlan string) error {
	apiToken := c.apiConfig.DefaultHeader["Authorization"]
	httpClient := &http.Client{}

	var advancedSettingsStrFromPlan string
	if advancedSettingsJsonFromPlan == "" {
		advancedSettingsStrFromPlan = "{}"
	} else {
		advancedSettingsStrFromPlan = advancedSettingsJsonFromPlan
	}

	//
	// Get service advanced settings
	urlAdvancedSettings, err := c.computeServiceAdvancedSettingsUrl(serviceType, serviceId)
	if err != nil {
		return err
	}

	overriddenAdvancedSettingsHashMap := make(map[string]any)
	if err := json.Unmarshal([]byte(advancedSettingsStrFromPlan), &overriddenAdvancedSettingsHashMap); err != nil {
		return err
	}

	getRequest, err := http.NewRequest("GET", *urlAdvancedSettings, nil)
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
		return errors.New("Cannot get advanced settings :" + respGetAdvancedSettings.Status)
	}
	serviceAdvancedSettings, err := io.ReadAll(respGetAdvancedSettings.Body)
	if err != nil {
		return err
	}

	advancedSettingsStringJson := string(serviceAdvancedSettings)

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
	putRequest, err := http.NewRequest(http.MethodPut, *urlAdvancedSettings, bytes.NewBuffer(overriddenAdvancedSettingsJSON))
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
		return errors.New("Cannot update service advanced settings :" + string(body))
	}

	return nil
}

// fetchDefaultAdvancedSettings fetches and parses the default advanced settings for a service
// type. These represent the full set of valid keys for that service type.
func (c ServiceAdvancedSettingsService) fetchDefaultAdvancedSettings(serviceType int) (map[string]any, error) {
	defaultAdvancedSettingsUrl, err := c.computeDefaultServiceAdvancedSettingsUrl(serviceType)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", *defaultAdvancedSettingsUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.apiConfig.DefaultHeader["Authorization"])
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.apiConfig.UserAgent)

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, errors.New("Cannot get default advanced settings :" + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	defaults := make(map[string]any)
	if err := json.Unmarshal(body, &defaults); err != nil {
		return nil, err
	}
	return defaults, nil
}

// defaultSettingKeys returns the set of valid advanced setting keys for a service type,
// caching the result per service type.
func (c ServiceAdvancedSettingsService) defaultSettingKeys(serviceType int) (map[string]struct{}, error) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	if cached, ok := c.defaultKeysCache[serviceType]; ok {
		return cached, nil
	}

	defaults, err := c.fetchDefaultAdvancedSettings(serviceType)
	if err != nil {
		return nil, err
	}

	keys := make(map[string]struct{}, len(defaults))
	for k := range defaults {
		keys[k] = struct{}{}
	}
	c.defaultKeysCache[serviceType] = keys
	return keys, nil
}

// computeUnknownKeys returns the keys in advancedSettings that are absent from validKeys,
// sorted for deterministic output.
func computeUnknownKeys(validKeys map[string]struct{}, advancedSettings map[string]any) []string {
	unknown := make([]string, 0)
	for key := range advancedSettings {
		if _, ok := validKeys[key]; !ok {
			unknown = append(unknown, key)
		}
	}
	sort.Strings(unknown)
	return unknown
}

// UnknownSettingKeys returns the advanced setting keys present in advancedSettingsJson that are
// not valid for the given service type (absent from that type's default settings). It returns
// nil for an empty input.
func (c ServiceAdvancedSettingsService) UnknownSettingKeys(serviceType int, advancedSettingsJson string) ([]string, error) {
	if advancedSettingsJson == "" || advancedSettingsJson == "{}" {
		return nil, nil
	}

	validKeys, err := c.defaultSettingKeys(serviceType)
	if err != nil {
		return nil, err
	}

	provided := make(map[string]any)
	if err := json.Unmarshal([]byte(advancedSettingsJson), &provided); err != nil {
		return nil, err
	}

	return computeUnknownKeys(validKeys, provided), nil
}
