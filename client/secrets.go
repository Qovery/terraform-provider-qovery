package client

import (
	"github.com/qovery/qovery-client-go"
)

type SecretsDiff struct {
	Create []SecretCreateRequest
	Update []SecretUpdateRequest
	Delete []SecretDeleteRequest
}

func (d SecretsDiff) IsEmpty() bool {
	return len(d.Create) == 0 &&
		len(d.Update) == 0 &&
		len(d.Delete) == 0
}

type SecretCreateRequest struct {
	qovery.SecretRequest
}

type SecretUpdateRequest struct {
	qovery.SecretEditRequest
	Id string
}

type SecretDeleteRequest struct {
	Id string
}

func secretResponseListToArray(list *qovery.SecretResponseList, scope qovery.APIVariableScopeEnum) []*qovery.Secret {
	vars := make([]*qovery.Secret, 0, len(list.GetResults()))
	for _, v := range list.GetResults() {
		if v.Scope != scope && v.Scope != qovery.APIVARIABLESCOPEENUM_BUILT_IN {
			continue
		}
		cpy := v
		vars = append(vars, &cpy)
	}
	return vars
}
