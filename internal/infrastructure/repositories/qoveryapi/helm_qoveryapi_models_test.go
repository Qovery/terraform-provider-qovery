package qoveryapi

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/helm"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

func TestNewDomainHelmFromQovery_NilResponse(t *testing.T) {
	t.Parallel()

	result, err := newDomainHelmFromQovery(nil, "deployment-stage-id", "{}", &qovery.CustomDomainResponseList{})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, variable.ErrNilVariable)
}

func TestNewQoveryHelmRequestFromDomain(t *testing.T) {
	t.Parallel()

	description := gofakeit.Sentence(3)
	timeout := int32(600)
	branch := "main"

	testCases := []struct {
		TestName    string
		Request     helm.UpsertRepositoryRequest
		ExpectError bool
	}{
		{
			TestName: "success_with_git_source",
			Request: helm.UpsertRepositoryRequest{
				Name:        "test-helm",
				Description: &description,
				AutoDeploy:  true,
				TimeoutSec:  &timeout,
				Arguments:   []string{"--wait", "--timeout=600s"},
				Source: helm.Source{
					GitRepository: &helm.SourceGitRepository{
						Url:      "https://github.com/example/helm-charts.git",
						Branch:   &branch,
						RootPath: "/charts/myapp",
					},
				},
				ValuesOverride: helm.ValuesOverride{
					Set:       [][]string{{"key1", "value1"}},
					SetString: [][]string{{"strKey", "strValue"}},
					SetJson:   [][]string{{"jsonKey", `{"nested":"value"}`}},
				},
				Ports:                     nil,
				AllowClusterWideResources: true,
			},
		},
		{
			TestName: "success_with_helm_repository_source",
			Request: helm.UpsertRepositoryRequest{
				Name:       "test-helm-repo",
				AutoDeploy: false,
				Source: helm.Source{
					HelmRepository: &helm.SourceHelmRepository{
						RepositoryId: gofakeit.UUID(),
						ChartName:    "nginx",
						ChartVersion: "1.0.0",
					},
				},
				ValuesOverride: helm.ValuesOverride{},
			},
		},
		{
			TestName: "success_with_ports",
			Request: helm.UpsertRepositoryRequest{
				Name:       "test-helm-with-ports",
				AutoDeploy: true,
				Source: helm.Source{
					GitRepository: &helm.SourceGitRepository{
						Url:      "https://github.com/example/helm-charts.git",
						Branch:   &branch,
						RootPath: "/",
					},
				},
				ValuesOverride: helm.ValuesOverride{},
				Ports: &[]helm.Port{
					{
						Name:         "http",
						InternalPort: 8080,
						ExternalPort: func() *int32 { p := int32(80); return &p }(),
						ServiceName:  "my-service",
						Protocol:     helm.ProtocolHttp,
						IsDefault:    true,
					},
					{
						Name:         "grpc",
						InternalPort: 9090,
						ServiceName:  "my-grpc-service",
						Protocol:     helm.ProtocolGrpc,
						IsDefault:    false,
					},
				},
			},
		},
		{
			TestName: "success_with_raw_values_override",
			Request: helm.UpsertRepositoryRequest{
				Name:       "test-helm-raw-values",
				AutoDeploy: true,
				Source: helm.Source{
					GitRepository: &helm.SourceGitRepository{
						Url:      "https://github.com/example/helm-charts.git",
						Branch:   &branch,
						RootPath: "/",
					},
				},
				ValuesOverride: helm.ValuesOverride{
					File: &helm.ValuesOverrideFile{
						Raw: &helm.Raw{
							Values: []helm.RawValue{
								{Name: "values.yaml", Content: "replicaCount: 3\nimage:\n  tag: latest"},
								{Name: "values-override.yaml", Content: "resources:\n  limits:\n    cpu: 1000m"},
							},
						},
					},
				},
			},
		},
		{
			TestName: "success_with_git_values_override",
			Request: helm.UpsertRepositoryRequest{
				Name:       "test-helm-git-values",
				AutoDeploy: true,
				Source: helm.Source{
					GitRepository: &helm.SourceGitRepository{
						Url:      "https://github.com/example/helm-charts.git",
						Branch:   &branch,
						RootPath: "/",
					},
				},
				ValuesOverride: helm.ValuesOverride{
					File: &helm.ValuesOverrideFile{
						GitRepository: &helm.ValuesOverrideGit{
							Url:    "https://github.com/example/values-repo.git",
							Branch: "main",
							Paths:  []string{"/values/prod.yaml", "/values/common.yaml"},
						},
					},
				},
			},
		},
		{
			TestName: "error_invalid_port_protocol",
			Request: helm.UpsertRepositoryRequest{
				Name:       "test-helm-invalid-port",
				AutoDeploy: true,
				Source: helm.Source{
					GitRepository: &helm.SourceGitRepository{
						Url:      "https://github.com/example/helm-charts.git",
						Branch:   &branch,
						RootPath: "/",
					},
				},
				ValuesOverride: helm.ValuesOverride{},
				Ports: &[]helm.Port{
					{
						Name:         "invalid",
						InternalPort: 8080,
						ServiceName:  "my-service",
						Protocol:     helm.Protocol("INVALID_PROTOCOL"),
						IsDefault:    true,
					},
				},
			},
			ExpectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			result, err := newQoveryHelmRequestFromDomain(tc.Request)
			if tc.ExpectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tc.Request.Name, result.Name)
			assert.Equal(t, tc.Request.AutoDeploy, result.AutoDeploy)
		})
	}
}

func TestNewQoveryHelmSourceRequestFromDomain(t *testing.T) {
	t.Parallel()

	branch := "main"
	developBranch := "develop"

	testCases := []struct {
		TestName             string
		Source               helm.Source
		ExpectGitRepository  bool
		ExpectHelmRepository bool
	}{
		{
			TestName: "git_repository_source",
			Source: helm.Source{
				GitRepository: &helm.SourceGitRepository{
					Url:      "https://github.com/example/charts.git",
					Branch:   &branch,
					RootPath: "/charts",
				},
			},
			ExpectGitRepository:  true,
			ExpectHelmRepository: false,
		},
		{
			TestName: "helm_repository_source",
			Source: helm.Source{
				HelmRepository: &helm.SourceHelmRepository{
					RepositoryId: gofakeit.UUID(),
					ChartName:    "myapp",
					ChartVersion: "2.0.0",
				},
			},
			ExpectGitRepository:  false,
			ExpectHelmRepository: true,
		},
		{
			TestName: "git_repository_with_token",
			Source: helm.Source{
				GitRepository: &helm.SourceGitRepository{
					Url:        "https://github.com/private/charts.git",
					Branch:     &developBranch,
					RootPath:   "/",
					GitTokenId: func() *string { s := gofakeit.UUID(); return &s }(),
				},
			},
			ExpectGitRepository:  true,
			ExpectHelmRepository: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			result := newQoveryHelmSourceRequestFromDomain(tc.Source)
			assert.NotNil(t, result)

			if tc.ExpectGitRepository {
				assert.NotNil(t, result.HelmRequestAllOfSourceOneOf)
				assert.NotNil(t, result.HelmRequestAllOfSourceOneOf.GitRepository)
			} else {
				assert.Nil(t, result.HelmRequestAllOfSourceOneOf)
			}

			if tc.ExpectHelmRepository {
				assert.NotNil(t, result.HelmRequestAllOfSourceOneOf1)
				assert.NotNil(t, result.HelmRequestAllOfSourceOneOf1.HelmRepository)
			} else {
				assert.Nil(t, result.HelmRequestAllOfSourceOneOf1)
			}
		})
	}
}

func TestNewQoveryHelmPortsRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Ports       *[]helm.Port
		ExpectedLen int
		ExpectError bool
	}{
		{
			TestName:    "nil_ports",
			Ports:       nil,
			ExpectedLen: 0,
		},
		{
			TestName:    "empty_ports",
			Ports:       &[]helm.Port{},
			ExpectedLen: 0,
		},
		{
			TestName: "single_http_port",
			Ports: &[]helm.Port{
				{
					Name:         "http",
					InternalPort: 8080,
					ExternalPort: func() *int32 { p := int32(80); return &p }(),
					ServiceName:  "web",
					Protocol:     helm.ProtocolHttp,
					IsDefault:    true,
				},
			},
			ExpectedLen: 1,
		},
		{
			TestName: "multiple_ports",
			Ports: &[]helm.Port{
				{
					Name:         "http",
					InternalPort: 8080,
					ServiceName:  "web",
					Protocol:     helm.ProtocolHttp,
					IsDefault:    true,
				},
				{
					Name:         "grpc",
					InternalPort: 9090,
					ServiceName:  "api",
					Protocol:     helm.ProtocolGrpc,
					IsDefault:    false,
				},
			},
			ExpectedLen: 2,
		},
		{
			TestName: "invalid_protocol",
			Ports: &[]helm.Port{
				{
					Name:         "invalid",
					InternalPort: 8080,
					ServiceName:  "test",
					Protocol:     helm.Protocol("INVALID"),
					IsDefault:    false,
				},
			},
			ExpectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			result, err := newQoveryHelmPortsRequestFromDomain(tc.Ports)
			if tc.ExpectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Len(t, *result, tc.ExpectedLen)
		})
	}
}

func TestNewQoveryFileValuesOverrideRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		File        *helm.ValuesOverrideFile
		ExpectError bool
	}{
		{
			TestName: "nil_file",
			File:     nil,
		},
		{
			TestName: "empty_file",
			File:     &helm.ValuesOverrideFile{},
		},
		{
			TestName: "raw_values",
			File: &helm.ValuesOverrideFile{
				Raw: &helm.Raw{
					Values: []helm.RawValue{
						{Name: "values.yaml", Content: "key: value"},
					},
				},
			},
		},
		{
			TestName: "git_values",
			File: &helm.ValuesOverrideFile{
				GitRepository: &helm.ValuesOverrideGit{
					Url:    "https://github.com/example/values.git",
					Branch: "main",
					Paths:  []string{"/values.yaml"},
				},
			},
		},
		{
			TestName: "both_raw_and_git",
			File: &helm.ValuesOverrideFile{
				Raw: &helm.Raw{
					Values: []helm.RawValue{
						{Name: "extra.yaml", Content: "extra: true"},
					},
				},
				GitRepository: &helm.ValuesOverrideGit{
					Url:    "https://github.com/example/values.git",
					Branch: "main",
					Paths:  []string{"/values.yaml"},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			result, err := newQoveryFileValuesOverrideRequestFromDomain(tc.File)
			if tc.ExpectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}
