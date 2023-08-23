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

type ServiceAdvancedSettingsService struct {
	apiConfig *qovery.Configuration
}

func NewServiceAdvancedSettingsService(apiConfig *qovery.Configuration) *ServiceAdvancedSettingsService {
	return &ServiceAdvancedSettingsService{apiConfig: apiConfig}
}

const (
	APPLICATION int = 0
	CONTAINER   int = 1
	JOB         int = 2
)

// Compute the URL to GET or PUT advanced settings for any service type
func (c ServiceAdvancedSettingsService) computeServiceAdvancedSettingsUrl(serviceType int, serviceId string) (*string, error) {
	var host = c.apiConfig.Servers[0].URL
	var urlAdvancedSettings string

	switch serviceType {
	case APPLICATION:
		urlAdvancedSettings = host + "/application/" + serviceId + "/advancedSettings"
	case CONTAINER:
		urlAdvancedSettings = host + "/container/" + serviceId + "/advancedSettings"
	case JOB:
		urlAdvancedSettings = host + "/job/" + serviceId + "/advancedSettings"
	default:
		return nil, errors.New("serviceType must be one of APPLICATION / CONTAINER / JOB")
	}

	return &urlAdvancedSettings, nil
}

// Compute the URL to GET default advanced settings for any service type
func (c ServiceAdvancedSettingsService) computeDefaultServiceAdvancedSettingsUrl(serviceType int) (*string, error) {
	var host = c.apiConfig.Servers[0].URL
	var urlAdvancedSettings string

	switch serviceType {
	case APPLICATION:
		urlAdvancedSettings = host + "/defaultApplicationAdvancedSettings"
	case CONTAINER:
		urlAdvancedSettings = host + "/defaultContainerAdvancedSettings"
	case JOB:
		urlAdvancedSettings = host + "/defaultJobAdvancedSettings"
	default:
		return nil, errors.New("serviceType must be one of APPLICATION / CONTAINER / JOB")
	}

	return &urlAdvancedSettings, nil
}

// ReadServiceAdvancedSettings Get only overridden advanced settings
func (c ServiceAdvancedSettingsService) ReadServiceAdvancedSettings(serviceType int, serviceId string) (*string, error) {
	httpClient := &http.Client{}
	var apiToken = c.apiConfig.DefaultHeader["Authorization"]

	//
	// Get default service advanced settings
	defaultAdvancedSettingsUrl, err := c.computeDefaultServiceAdvancedSettingsUrl(serviceType)
	if err != nil {
		return nil, err
	}
	getDefaultAdvancedSettingsRequest, err := http.NewRequest("GET", *defaultAdvancedSettingsUrl, nil)
	if err != nil {
		return nil, err
	}
	getDefaultAdvancedSettingsRequest.Header.Set("Authorization", apiToken)
	getDefaultAdvancedSettingsRequest.Header.Set("Content-Type", "application/json")

	respGetDefaultAdvancedSettings, err := httpClient.Do(getDefaultAdvancedSettingsRequest)
	defer respGetDefaultAdvancedSettings.Body.Close()

	if err != nil || respGetDefaultAdvancedSettings.StatusCode >= 400 {
		return nil, errors.New("Cannot get default advanced settings :" + respGetDefaultAdvancedSettings.Status)
	}

	serviceDefaultAdvancedSettings, err := io.ReadAll(respGetDefaultAdvancedSettings.Body)
	if err != nil {
		return nil, err
	}

	defaultAdvancedSettingsStringJson := string(serviceDefaultAdvancedSettings)

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

	respGetAdvancedSettings, err := httpClient.Do(getRequest)
	defer respGetAdvancedSettings.Body.Close()

	if err != nil || respGetAdvancedSettings.StatusCode >= 400 {
		return nil, errors.New("Cannot get advanced settings :" + respGetAdvancedSettings.Status)
	}

	serviceAdvancedSettings, err := io.ReadAll(respGetAdvancedSettings.Body)
	if err != nil {
		return nil, err
	}

	advancedSettingsStringJson := string(serviceAdvancedSettings)

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
		return nil, errors.New("Cannot parse overridden advanced settings")
	}

	s := string(overridenAdvancedSettingsJson)
	return &s, nil
}

// UpdateServiceAdvancedSettings Update advanced settings by computing the whole http body
func (c ServiceAdvancedSettingsService) UpdateServiceAdvancedSettings(serviceType int, serviceId string, advancedSettingsJson string) error {
	var apiToken = c.apiConfig.DefaultHeader["Authorization"]
	httpClient := &http.Client{}

	//
	// Get service advanced settings
	urlAdvancedSettings, err := c.computeServiceAdvancedSettingsUrl(serviceType, serviceId)
	if err != nil {
		return err
	}

	overridenAdvancedSettingsHashMap := make(map[string]interface{})
	json.Unmarshal([]byte(advancedSettingsJson), &overridenAdvancedSettingsHashMap)

	getRequest, err := http.NewRequest("GET", *urlAdvancedSettings, nil)
	if err != nil {
		return err
	}
	getRequest.Header.Set("Authorization", apiToken)
	getRequest.Header.Set("Content-Type", "application/json")

	respGetAdvancedSettings, err := httpClient.Do(getRequest)
	defer respGetAdvancedSettings.Body.Close()

	if err != nil || respGetAdvancedSettings.StatusCode >= 400 {
		return errors.New("Cannot get advanced settings :" + respGetAdvancedSettings.Status)
	}
	serviceAdvancedSettings, err := io.ReadAll(respGetAdvancedSettings.Body)
	if err != nil {
		return err
	}

	advancedSettingsStringJson := string(serviceAdvancedSettings)

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
	putRequest, err := http.NewRequest(http.MethodPut, *urlAdvancedSettings, bytes.NewBuffer(overridenAdvancedSettingsJson))
	if err != nil {
		return err
	}
	putRequest.Header.Set("Authorization", apiToken)
	putRequest.Header.Set("Content-Type", "application/json")
	putRequest.Header.Set("Accept", "application/json")

	respPostAdvancedSettings, err := httpClient.Do(putRequest)

	defer respPostAdvancedSettings.Body.Close()

	if err != nil || respPostAdvancedSettings.StatusCode >= 400 {
		return err
	}

	return nil
}