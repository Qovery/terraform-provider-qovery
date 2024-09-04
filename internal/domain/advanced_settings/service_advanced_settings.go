package advanced_settings

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain"
)

type ServiceAdvancedSettingsService struct {
	apiConfig *qovery.Configuration
}

func NewServiceAdvancedSettingsService(apiConfig *qovery.Configuration) *ServiceAdvancedSettingsService {
	return &ServiceAdvancedSettingsService{apiConfig: apiConfig}
}

// Compute the URL to GET or PUT advanced settings for any service type
func (c ServiceAdvancedSettingsService) computeServiceAdvancedSettingsUrl(serviceType int, serviceId string) (*string, error) {
	var host = c.apiConfig.Servers[0].URL
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
	default:
		return nil, errors.New("serviceType must be one of APPLICATION / CONTAINER / JOB / HELM")
	}

	return &urlAdvancedSettings, nil
}

// Compute the URL to GET default advanced settings for any service type
func (c ServiceAdvancedSettingsService) computeDefaultServiceAdvancedSettingsUrl(serviceType int) (*string, error) {
	var host = c.apiConfig.Servers[0].URL
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
	default:
		return nil, errors.New("serviceType must be one of APPLICATION / CONTAINER / JOB / HELM")
	}

	return &urlAdvancedSettings, nil
}

// ReadServiceAdvancedSettings Get only overridden advanced settings
func (c ServiceAdvancedSettingsService) ReadServiceAdvancedSettings(serviceType int, serviceId string, advancedSettingsJsonFromState string) (*string, error) {
	httpClient := &http.Client{}
	var apiToken = c.apiConfig.DefaultHeader["Authorization"]

	var serviceAdvancedSettingsState string
	if advancedSettingsJsonFromState == "" {
		serviceAdvancedSettingsState = "{}"
	} else {
		serviceAdvancedSettingsState = advancedSettingsJsonFromState
	}

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
	getDefaultAdvancedSettingsRequest.Header.Set("User-Agent", c.apiConfig.UserAgent)

	respGetDefaultAdvancedSettings, err := httpClient.Do(getDefaultAdvancedSettingsRequest)
	if err != nil {
		return nil, err
	}
	defer respGetDefaultAdvancedSettings.Body.Close()

	if respGetDefaultAdvancedSettings.StatusCode >= 400 {
		return nil, errors.New("Cannot get default advanced settings :" + respGetDefaultAdvancedSettings.Status)
	}

	defaultServiceAdvancedSettingsJson, err := io.ReadAll(respGetDefaultAdvancedSettings.Body)
	if err != nil {
		return nil, err
	}

	defaultAdvancedSettingsJsonString := string(defaultServiceAdvancedSettingsJson)

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
	advancedSettingsFromStateHashMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(serviceAdvancedSettingsState), &advancedSettingsFromStateHashMap); err != nil {
		return nil, err
	}

	currentAdvancedSettingsHashMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(advancedSettingsStringJson), &currentAdvancedSettingsHashMap); err != nil {
		return nil, err
	}

	defaultAdvancedSettingsHashMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(defaultAdvancedSettingsJsonString), &defaultAdvancedSettingsHashMap); err != nil {
		return nil, err
	}

	overriddenAdvancedSettings := make(map[string]interface{})
	// Prepare hashmap with target advanced settings
	for advanced_setting_name, advanced_setting_value := range currentAdvancedSettingsHashMap {
		defaultValue := defaultAdvancedSettingsHashMap[advanced_setting_name]
		// if the value is not in the state ignore it
		// otherwise if an advanced setting has been modified in the UI we don't want to show the diff
		_, ok := advancedSettingsFromStateHashMap[advanced_setting_name]
		if !ok {
			continue
		}
		// if the value has been overridden
		if !reflect.DeepEqual(defaultValue, advanced_setting_value) {
			overriddenAdvancedSettings[advanced_setting_name] = advanced_setting_value
		} else {
			// if the value is in the state
			v, ok := advancedSettingsFromStateHashMap[advanced_setting_name]
			if ok {
				overriddenAdvancedSettings[advanced_setting_name] = v
			}
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
func (c ServiceAdvancedSettingsService) UpdateServiceAdvancedSettings(serviceType int, serviceId string, advancedSettingsJsonFromPlan string) error {
	var apiToken = c.apiConfig.DefaultHeader["Authorization"]
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

	overridenAdvancedSettingsHashMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(advancedSettingsStrFromPlan), &overridenAdvancedSettingsHashMap); err != nil {
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
	if err := json.Unmarshal([]byte(advancedSettingsStringJson), &currentAdvancedSettingsHashMap); err != nil {
		return err
	}

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
