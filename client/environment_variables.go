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

func environmentVariableResponseListToArray(list *qovery.EnvironmentVariableResponseList, scope qovery.EnvironmentVariableScopeEnum) []*qovery.EnvironmentVariable {
	vars := make([]*qovery.EnvironmentVariable, 0, len(list.GetResults()))
	for _, v := range list.GetResults() {
		if v.Scope != scope && v.Scope != qovery.ENVIRONMENTVARIABLESCOPEENUM_BUILT_IN {
			continue
		}
		cpy := v
		vars = append(vars, &cpy)
	}
	return vars
}
