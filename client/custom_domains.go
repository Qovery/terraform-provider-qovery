package client

import (
	"github.com/qovery/qovery-client-go"
)

type CustomDomainsDiff struct {
	Create []CustomDomainCreateRequest
	Update []CustomDomainUpdateRequest
	Delete []CustomDomainDeleteRequest
}

func (d CustomDomainsDiff) IsEmpty() bool {
	return len(d.Create) == 0 &&
		len(d.Update) == 0 &&
		len(d.Delete) == 0
}

type CustomDomainCreateRequest struct {
	qovery.CustomDomainRequest
}

type CustomDomainUpdateRequest struct {
	qovery.CustomDomainRequest
	Id string
}

type CustomDomainDeleteRequest struct {
	Id string
}

func customDomainResponseListToArray(list *qovery.CustomDomainResponseList) []*qovery.CustomDomain {
	vars := make([]*qovery.CustomDomain, 0, len(list.GetResults()))
	for _, v := range list.GetResults() {
		cpy := v
		vars = append(vars, &cpy)
	}
	return vars
}
