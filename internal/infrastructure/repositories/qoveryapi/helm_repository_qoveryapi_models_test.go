package qoveryapi

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/helmRepository"
	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
)

func TestNewDomainHelmRepositoryFromQovery(t *testing.T) {
	t.Parallel()

	description := gofakeit.Sentence(3)
	skipTLS := true
	kindHTTPS := qovery.HELMREPOSITORYKINDENUM_HTTPS
	kindOCIECR := qovery.HELMREPOSITORYKINDENUM_OCI_ECR
	url1 := "https://charts.example.com"
	url2 := "https://123456789.dkr.ecr.us-east-1.amazonaws.com"

	testCases := []struct {
		TestName       string
		Response       *qovery.HelmRepositoryResponse
		OrganizationID string
		ExpectedError  error
	}{
		{
			TestName:       "error_with_nil_response",
			Response:       nil,
			OrganizationID: gofakeit.UUID(),
			ExpectedError:  registry.ErrNilRegistry,
		},
		{
			TestName: "success",
			Response: &qovery.HelmRepositoryResponse{
				Id:                  gofakeit.UUID(),
				Name:                gofakeit.Word(),
				Kind:                &kindHTTPS,
				Url:                 &url1,
				Description:         &description,
				SkipTlsVerification: &skipTLS,
			},
			OrganizationID: gofakeit.UUID(),
		},
		{
			TestName: "success_with_oci_ecr_kind",
			Response: &qovery.HelmRepositoryResponse{
				Id:   gofakeit.UUID(),
				Name: gofakeit.Word(),
				Kind: &kindOCIECR,
				Url:  &url2,
			},
			OrganizationID: gofakeit.UUID(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			result, err := newDomainHelmRepositoryFromQovery(tc.Response, tc.OrganizationID)
			if tc.ExpectedError != nil {
				assert.ErrorIs(t, err, tc.ExpectedError)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tc.Response.Id, result.ID.String())
			assert.Equal(t, tc.OrganizationID, result.OrganizationID.String())
			assert.Equal(t, tc.Response.Name, result.Name)
			assert.Equal(t, string(*tc.Response.Kind), result.Kind.String())
			assert.Equal(t, *tc.Response.Url, result.URL.String())
		})
	}
}

func TestNewQoveryHelmRepositoryRequestFromDomain(t *testing.T) {
	t.Parallel()

	description := gofakeit.Sentence(3)

	testCases := []struct {
		TestName      string
		Request       helmRepository.UpsertRequest
		ExpectedError error
	}{
		{
			TestName: "error_with_invalid_kind",
			Request: helmRepository.UpsertRequest{
				Name: gofakeit.Word(),
				Kind: "INVALID_KIND",
				URL:  "https://charts.example.com",
			},
			ExpectedError: registry.ErrInvalidKindParam,
		},
		{
			TestName: "success_with_https",
			Request: helmRepository.UpsertRequest{
				Name:               gofakeit.Word(),
				Kind:               "HTTPS",
				URL:                "https://charts.example.com",
				Description:        &description,
				SkiTlsVerification: true,
				Config: registry.UpsertRequestConfig{
					Username: func() *string { s := gofakeit.Username(); return &s }(),
					Password: func() *string { s := gofakeit.Password(true, true, true, false, false, 10); return &s }(),
				},
			},
		},
		{
			TestName: "success_with_oci_ecr",
			Request: helmRepository.UpsertRequest{
				Name: gofakeit.Word(),
				Kind: "OCI_ECR",
				URL:  "https://123456789.dkr.ecr.us-east-1.amazonaws.com",
				Config: registry.UpsertRequestConfig{
					AccessKeyID:     func() *string { s := gofakeit.Word(); return &s }(),
					SecretAccessKey: func() *string { s := gofakeit.Word(); return &s }(),
					Region:          func() *string { s := "us-east-1"; return &s }(),
				},
			},
		},
		{
			TestName: "success_with_oci_scaleway_cr",
			Request: helmRepository.UpsertRequest{
				Name: gofakeit.Word(),
				Kind: "OCI_SCALEWAY_CR",
				URL:  "https://rg.fr-par.scw.cloud",
				Config: registry.UpsertRequestConfig{
					ScalewayAccessKey: func() *string { s := gofakeit.Word(); return &s }(),
					ScalewaySecretKey: func() *string { s := gofakeit.Word(); return &s }(),
					ScalewayProjectId: func() *string { s := gofakeit.UUID(); return &s }(),
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			result, err := newQoveryHelmRepositoryRequestFromDomain(tc.Request)
			if tc.ExpectedError != nil {
				assert.ErrorIs(t, err, tc.ExpectedError)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tc.Request.Name, result.Name)
			assert.Equal(t, tc.Request.Kind, string(result.Kind))
			assert.Equal(t, tc.Request.URL, *result.Url)
		})
	}
}
