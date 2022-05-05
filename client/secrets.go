package client

import (
	"github.com/qovery/qovery-client-go"
)

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
	qovery.EnvironmentVariableRequest
}

type EnvironmentVariableUpdateRequest struct {
	qovery.EnvironmentVariableEditRequest
	Id string
}

type EnvironmentVariableDeleteRequest struct {
	Id string
}
