//go:build unit && !integration
// +build unit,!integration

package qoveryapi

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
)

func TestNewDomainHelmRepositoryFromQovery_Config(t *testing.T) {
	t.Parallel()

	response := &qovery.HelmRepositoryResponse{
		Id:   gofakeit.UUID(),
		Name: gofakeit.Name(),
		Kind: qovery.HELMREPOSITORYKINDENUM_OCI_ECR.Ptr(),
		Url:  new("oci://123456789012.dkr.ecr.eu-west-3.amazonaws.com"),
		Config: &qovery.HelmRepositoryResponseAllOfConfig{
			Username:          new("user"),
			Region:            new("eu-west-3"),
			AccessKeyId:       new("AKIA"),
			RoleArn:           new("arn:aws:iam::123456789012:role/qovery"),
			ScalewayAccessKey: new("SCW"),
			ScalewayProjectId: new("scaleway-project"),
		},
	}

	repository, err := newDomainHelmRepositoryFromQovery(response, gofakeit.UUID())
	require.NoError(t, err)
	assert.Equal(t, registry.Config{
		AccessKeyID:       new("AKIA"),
		Region:            new("eu-west-3"),
		ScalewayAccessKey: new("SCW"),
		ScalewayProjectID: new("scaleway-project"),
		Username:          new("user"),
	}, repository.Config)

	response.Config = nil
	repository, err = newDomainHelmRepositoryFromQovery(response, gofakeit.UUID())
	require.NoError(t, err)
	assert.Equal(t, registry.Config{}, repository.Config, "a response without config has no key")
}
