package client

import (
	"github.com/qovery/qovery-client-go"
)

type EnvironmentVariable struct {
	Key   string
	Value string
	Scope qovery.EnvironmentVariableScopeEnum
}

type EnvironmentVariablesDiff struct {
	Create []EnvironmentVariableCreateRequest
	Update []EnvironmentVariableUpdateRequest
	Delete []EnvironmentVariableDeleteRequest
}

func (d EnvironmentVariablesDiff) IsEmpty() bool {
	return len(d.Create) == 0 &&
		len(d.Update) == 0 &&
		len(d.Delete) == 0
}

type EnvironmentVariableCreateRequest struct {
	EnvironmentVariable
}

func (e EnvironmentVariableCreateRequest) toRequest() qovery.EnvironmentVariableRequest {
	return qovery.EnvironmentVariableRequest{
		Key:   e.Key,
		Value: e.Value,
	}
}

type EnvironmentVariableUpdateRequest struct {
	EnvironmentVariable
	Id string
}

func (e EnvironmentVariableUpdateRequest) toRequest() qovery.EnvironmentVariableEditRequest {
	return qovery.EnvironmentVariableEditRequest{
		Key:   e.Key,
		Value: e.Value,
	}
}

type EnvironmentVariableDeleteRequest struct {
	EnvironmentVariable
	Id string
}

func environmentVariableResponseListToArray(list *qovery.EnvironmentVariableResponseList) []*qovery.EnvironmentVariable {
	vars := make([]*qovery.EnvironmentVariable, 0, len(list.GetResults()))
	for _, v := range list.GetResults() {
		cpy := v
		vars = append(vars, &cpy)
	}
	return vars
}
