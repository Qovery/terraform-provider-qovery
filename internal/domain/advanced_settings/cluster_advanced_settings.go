package advanced_settings

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"
)

type ClusterAdvancedSettingsService struct {
	apiConfig *qovery.Configuration
}

func NewClusterAdvancedSettingsService(apiConfig *qovery.Configuration) *ClusterAdvancedSettingsService {
	return &ClusterAdvancedSettingsService{apiConfig: apiConfig}
}

// ReadServiceAdvancedSettings Get only overridden advanced settings
func (c ClusterAdvancedSettingsService) ReadClusterAdvancedSettings(organizationId string, clusterId string) (*string, error) {
	httpClient := &http.Client{}
	var apiToken = c.apiConfig.DefaultHeader["Authorization"]
	var host = c.apiConfig.Servers[0].URL

	//
	// Get default cluster advanced settings
	defaultAdvancedSettingsUrl := host + "/defaultClusterAdvancedSettings"
	getDefaultAdvancedSettingsRequest, err := http.NewRequest("GET", defaultAdvancedSettingsUrl, nil)
	if err != nil {
		return nil, err
	}
	getDefaultAdvancedSettingsRequest.Header.Set("Authorization", apiToken)
	getDefaultAdvancedSettingsRequest.Header.Set("Content-Type", "application/json")
	getDefaultAdvancedSettingsRequest.Header.Set("User-Agent", c.apiConfig.UserAgent)

	respGetDefaultAdvancedSettings, err := httpClient.Do(getDefaultAdvancedSettingsRequest)
	defer respGetDefaultAdvancedSettings.Body.Close()

	if err != nil || respGetDefaultAdvancedSettings.StatusCode >= 400 {
		return nil, errors.New("Cannot get default cluster advanced settings :" + respGetDefaultAdvancedSettings.Status)
	}

	clusterDefaultAdvancedSettings, err := io.ReadAll(respGetDefaultAdvancedSettings.Body)
	if err != nil {
		return nil, err
	}

	defaultAdvancedSettingsStringJson := string(clusterDefaultAdvancedSettings)

	//
	// Get cluster advanced settings
	urlAdvancedSettings := host + "/organization/" + organizationId + "/cluster/" + clusterId + "/advancedSettings"
	getRequest, err := http.NewRequest("GET", urlAdvancedSettings, nil)
	if err != nil {
		return nil, err
	}
	getRequest.Header.Set("Authorization", apiToken)
	getRequest.Header.Set("Content-Type", "application/json")
	getRequest.Header.Set("User-Agent", c.apiConfig.UserAgent)

	respGetAdvancedSettings, err := httpClient.Do(getRequest)
	defer respGetAdvancedSettings.Body.Close()

	if err != nil || respGetAdvancedSettings.StatusCode >= 400 {
		return nil, errors.New("Cannot get cluster advanced settings :" + respGetAdvancedSettings.Status)
	}

	clusterAdvancedSettings, err := io.ReadAll(respGetAdvancedSettings.Body)
	if err != nil {
		return nil, err
	}

	advancedSettingsStringJson := string(clusterAdvancedSettings)

	//
	// Compute the Diff
	currentAdvancedSettingsHashMap := make(map[string]interface{})
	json.Unmarshal([]byte(advancedSettingsStringJson), &currentAdvancedSettingsHashMap)

	defaultAdvancedSettingsHashMap := make(map[string]interface{})
	json.Unmarshal([]byte(defaultAdvancedSettingsStringJson), &defaultAdvancedSettingsHashMap)

	overriddenAdvancedSettings := make(map[string]interface{})
	// Prepare hashmap with target advanced settings
	for k, v := range currentAdvancedSettingsHashMap {
		defaultValue, _ := defaultAdvancedSettingsHashMap[k]
		// if the value has been overridden
		if !reflect.DeepEqual(defaultValue, v) {
			overriddenAdvancedSettings[k] = v
		}
	}

	//
	// Transform to JSON
	overridenAdvancedSettingsJson, err := json.Marshal(overriddenAdvancedSettings)
	if err != nil {
		return nil, errors.New("Cannot parse overridden cluster advanced settings")
	}

	s := string(overridenAdvancedSettingsJson)
	return &s, nil
}

// UpdateServiceAdvancedSettings Update advanced settings by computing the whole http body
func (c ClusterAdvancedSettingsService) UpdateClusterAdvancedSettings(organizationId string, clusterId string, advancedSettingsJson string) error {
	httpClient := &http.Client{}
	var apiToken = c.apiConfig.DefaultHeader["Authorization"]
	var host = c.apiConfig.Servers[0].URL

	//
	// Get cluster advanced settings
	urlAdvancedSettings := host + "/organization/" + organizationId + "/cluster/" + clusterId + "/advancedSettings"
	overridenAdvancedSettingsHashMap := make(map[string]interface{})
	json.Unmarshal([]byte(advancedSettingsJson), &overridenAdvancedSettingsHashMap)

	getRequest, err := http.NewRequest("GET", urlAdvancedSettings, nil)
	if err != nil {
		return err
	}
	getRequest.Header.Set("Authorization", apiToken)
	getRequest.Header.Set("Content-Type", "application/json")
	getRequest.Header.Set("User-Agent", c.apiConfig.UserAgent)

	respGetAdvancedSettings, err := httpClient.Do(getRequest)
	defer respGetAdvancedSettings.Body.Close()

	if err != nil || respGetAdvancedSettings.StatusCode >= 400 {
		return errors.New("Cannot get cluster advanced settings :" + respGetAdvancedSettings.Status)
	}
	clusterAdvancedSettings, err := io.ReadAll(respGetAdvancedSettings.Body)
	if err != nil {
		return err
	}

	advancedSettingsStringJson := string(clusterAdvancedSettings)

	//
	// Compute final http body to send to satisfy PUT endpoint
	currentAdvancedSettingsHashMap := make(map[string]interface{})
	json.Unmarshal([]byte(advancedSettingsStringJson), &currentAdvancedSettingsHashMap)

	for k, v := range currentAdvancedSettingsHashMap {
		_, exists := overridenAdvancedSettingsHashMap[k]
		if !exists {
			overridenAdvancedSettingsHashMap[k] = v
		}
	}

	overridenAdvancedSettingsJson, err := json.Marshal(overridenAdvancedSettingsHashMap)
	if err != nil {
		return err
	}

	//
	// Update advanced settings
	putRequest, err := http.NewRequest(http.MethodPut, urlAdvancedSettings, bytes.NewBuffer(overridenAdvancedSettingsJson))
	if err != nil {
		return err
	}
	putRequest.Header.Set("Authorization", apiToken)
	putRequest.Header.Set("Content-Type", "application/json")
	putRequest.Header.Set("Accept", "application/json")
	putRequest.Header.Set("User-Agent", c.apiConfig.UserAgent)

	respPostAdvancedSettings, err := httpClient.Do(putRequest)

	defer respPostAdvancedSettings.Body.Close()

	if err != nil || respPostAdvancedSettings.StatusCode >= 400 {
		return err
	}

	return nil
}
