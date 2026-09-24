//go:build unit && !integration

package services

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/blueprint"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deployment"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

func newTestBlueprintService(repo blueprint.Repository) blueprintService {
	return blueprintService{
		blueprintRepository: repo,
		waitTimeout:         time.Second,
		waitPollInterval:    time.Millisecond,
	}
}

func testBlueprint(environmentID uuid.UUID, dispatch *blueprint.Dispatch, serviceID *string) *blueprint.Blueprint {
	return &blueprint.Blueprint{
		ID:               uuid.New(),
		EnvironmentID:    environmentID,
		Name:             "my-db",
		Tag:              "aws/postgres/17/1.0.0",
		ServiceType:      blueprint.ServiceTypeTerraform,
		ServiceID:        serviceID,
		LatestDeployment: dispatch,
	}
}

func strPtr(s string) *string { return &s }

func TestBlueprintServiceCreate(t *testing.T) {
	t.Parallel()
	environmentID := uuid.New()
	blueprintID := uuid.NewString()
	serviceID := uuid.NewString()
	validRequest := blueprint.CreateRequest{UpsertRequest: blueprint.UpsertRequest{Name: "my-db", Tag: "aws/postgres/17/1.0.0", IconURI: "app://qovery-console/terraform"}}
	created := &blueprint.CreationResult{BlueprintID: blueprintID, DispatchID: "dispatch-2"}

	t.Run("invalid environment id", func(t *testing.T) {
		svc := newTestBlueprintService(mocks_test.NewBlueprintRepository(t))
		bp, err := svc.Create(context.Background(), "not-a-uuid", validRequest)
		assert.Nil(t, bp)
		assert.ErrorContains(t, err, blueprint.ErrInvalidEnvironmentIDParam.Error())
	})

	t.Run("variable declared as secret and non-secret", func(t *testing.T) {
		svc := newTestBlueprintService(mocks_test.NewBlueprintRepository(t))
		request := validRequest
		request.Variables = map[string]string{"password": "a"}
		request.SecretVariables = map[string]string{"password": "b"}
		bp, err := svc.Create(context.Background(), environmentID.String(), request)
		assert.Nil(t, bp)
		assert.ErrorIs(t, err, blueprint.ErrVariableDeclaredTwice)
	})

	t.Run("repository create error returns no blueprint", func(t *testing.T) {
		repo := mocks_test.NewBlueprintRepository(t)
		repo.EXPECT().Create(mock.Anything, environmentID.String(), validRequest).Return(nil, errors.New("boom"))
		bp, err := newTestBlueprintService(repo).Create(context.Background(), environmentID.String(), validRequest)
		assert.Nil(t, bp)
		assert.ErrorContains(t, err, blueprint.ErrFailedToCreateBlueprint.Error())
	})

	t.Run("waits for its own dispatch then succeeds without deploying", func(t *testing.T) {
		repo := mocks_test.NewBlueprintRepository(t)
		repo.EXPECT().Create(mock.Anything, environmentID.String(), validRequest).Return(created, nil)
		earlier := testBlueprint(environmentID, &blueprint.Dispatch{ID: "dispatch-1", Status: blueprint.DispatchStatusRunning}, nil)
		pending := testBlueprint(environmentID, &blueprint.Dispatch{ID: "dispatch-2", Status: blueprint.DispatchStatusDeploying}, nil)
		done := testBlueprint(environmentID, &blueprint.Dispatch{ID: "dispatch-2", Status: blueprint.DispatchStatusRunning}, &serviceID)
		repo.EXPECT().Get(mock.Anything, blueprintID).Return(earlier, nil).Once()
		repo.EXPECT().Get(mock.Anything, blueprintID).Return(pending, nil).Once()
		repo.EXPECT().Get(mock.Anything, blueprintID).Return(done, nil).Once()

		bp, err := newTestBlueprintService(repo).Create(context.Background(), environmentID.String(), validRequest)
		require.NoError(t, err)
		assert.Equal(t, done, bp)
	})

	t.Run("read failure right after create still returns the blueprint id", func(t *testing.T) {
		repo := mocks_test.NewBlueprintRepository(t)
		repo.EXPECT().Create(mock.Anything, environmentID.String(), validRequest).Return(created, nil)
		repo.EXPECT().Get(mock.Anything, blueprintID).Return(nil, errors.New("502")).Once()

		bp, err := newTestBlueprintService(repo).Create(context.Background(), environmentID.String(), validRequest)
		assert.ErrorContains(t, err, blueprint.ErrFailedToCreateBlueprint.Error())
		require.NotNil(t, bp)
		assert.Equal(t, blueprintID, bp.ID.String())
		assert.Equal(t, environmentID, bp.EnvironmentID)
		assert.Equal(t, validRequest.Name, bp.Name)
	})

	t.Run("failed dispatch returns blueprint and engine message", func(t *testing.T) {
		repo := mocks_test.NewBlueprintRepository(t)
		repo.EXPECT().Create(mock.Anything, environmentID.String(), validRequest).Return(created, nil)
		failed := testBlueprint(environmentID, &blueprint.Dispatch{ID: "dispatch-2", Status: blueprint.DispatchStatusFailed, ErrorMessage: strPtr("terraform apply failed")}, nil)
		repo.EXPECT().Get(mock.Anything, blueprintID).Return(failed, nil).Once()

		bp, err := newTestBlueprintService(repo).Create(context.Background(), environmentID.String(), validRequest)
		assert.Equal(t, failed, bp)
		assert.ErrorIs(t, err, blueprint.ErrDispatchFailed)
		assert.ErrorContains(t, err, "terraform apply failed")
	})

	t.Run("successful dispatch without service is an error", func(t *testing.T) {
		repo := mocks_test.NewBlueprintRepository(t)
		repo.EXPECT().Create(mock.Anything, environmentID.String(), validRequest).Return(created, nil)
		done := testBlueprint(environmentID, &blueprint.Dispatch{ID: "dispatch-2", Status: blueprint.DispatchStatusRunning}, nil)
		repo.EXPECT().Get(mock.Anything, blueprintID).Return(done, nil).Once()

		bp, err := newTestBlueprintService(repo).Create(context.Background(), environmentID.String(), validRequest)
		assert.Equal(t, done, bp)
		assert.ErrorIs(t, err, blueprint.ErrDispatchDidNotCreateService)
	})

	dispatchStart := time.Date(2026, 9, 24, 14, 21, 46, 0, time.UTC)
	after := dispatchStart.Add(5 * time.Minute)

	t.Run("deploy waits for a service deployment newer than the dispatch", func(t *testing.T) {
		request := validRequest
		request.Deploy = true
		repo := mocks_test.NewBlueprintRepository(t)
		repo.EXPECT().Create(mock.Anything, environmentID.String(), request).Return(created, nil)
		done := testBlueprint(environmentID, &blueprint.Dispatch{ID: "dispatch-2", Status: blueprint.DispatchStatusRunning, StartedAt: dispatchStart}, &serviceID)
		repo.EXPECT().Get(mock.Anything, blueprintID).Return(done, nil).Once()
		repo.EXPECT().GetServiceStatus(mock.Anything, environmentID.String(), blueprint.ServiceTypeTerraform, serviceID).Return(nil, nil).Once()
		repo.EXPECT().GetServiceStatus(mock.Anything, environmentID.String(), blueprint.ServiceTypeTerraform, serviceID).Return(&blueprint.ServiceStatus{State: "READY"}, nil).Once()
		repo.EXPECT().GetServiceStatus(mock.Anything, environmentID.String(), blueprint.ServiceTypeTerraform, serviceID).Return(&blueprint.ServiceStatus{State: "QUEUED", LastDeploymentDate: &after}, nil).Once()
		repo.EXPECT().GetServiceStatus(mock.Anything, environmentID.String(), blueprint.ServiceTypeTerraform, serviceID).Return(&blueprint.ServiceStatus{State: "DEPLOYING", LastDeploymentDate: &after}, nil).Once()
		repo.EXPECT().GetServiceStatus(mock.Anything, environmentID.String(), blueprint.ServiceTypeTerraform, serviceID).Return(&blueprint.ServiceStatus{State: "DEPLOYED", LastDeploymentDate: &after}, nil).Once()

		bp, err := newTestBlueprintService(repo).Create(context.Background(), environmentID.String(), request)
		require.NoError(t, err)
		assert.Equal(t, done, bp)
	})

	t.Run("failed service deployment returns blueprint and error", func(t *testing.T) {
		request := validRequest
		request.Deploy = true
		repo := mocks_test.NewBlueprintRepository(t)
		repo.EXPECT().Create(mock.Anything, environmentID.String(), request).Return(created, nil)
		done := testBlueprint(environmentID, &blueprint.Dispatch{ID: "dispatch-2", Status: blueprint.DispatchStatusRunning, StartedAt: dispatchStart}, &serviceID)
		repo.EXPECT().Get(mock.Anything, blueprintID).Return(done, nil).Once()
		repo.EXPECT().GetServiceStatus(mock.Anything, environmentID.String(), blueprint.ServiceTypeTerraform, serviceID).Return(&blueprint.ServiceStatus{State: "DEPLOYMENT_ERROR", LastDeploymentDate: &after}, nil).Once()

		bp, err := newTestBlueprintService(repo).Create(context.Background(), environmentID.String(), request)
		assert.Equal(t, done, bp)
		assert.ErrorIs(t, err, blueprint.ErrServiceDeploymentFailed)
		assert.True(t, bp.LastApplyFailed(), "the returned blueprint must carry the failed service status")
	})

	t.Run("a dispatch that never ends times out with the blueprint sentinel", func(t *testing.T) {
		repo := mocks_test.NewBlueprintRepository(t)
		repo.EXPECT().Create(mock.Anything, environmentID.String(), validRequest).Return(created, nil)
		pending := testBlueprint(environmentID, &blueprint.Dispatch{ID: "dispatch-2", Status: blueprint.DispatchStatusDeploying}, nil)
		repo.EXPECT().Get(mock.Anything, blueprintID).Return(pending, nil)
		svc := newTestBlueprintService(repo)
		svc.waitTimeout = 20 * time.Millisecond

		_, err := svc.Create(context.Background(), environmentID.String(), validRequest)
		assert.ErrorIs(t, err, blueprint.ErrWaitTimeout)
		assert.NotErrorIs(t, err, deployment.ErrWaitTimeout)
	})
}

func TestBlueprintServiceUpdate(t *testing.T) {
	t.Parallel()
	environmentID := uuid.New()
	blueprintID := uuid.NewString()
	serviceID := uuid.NewString()
	request := blueprint.UpdateRequest{UpsertRequest: blueprint.UpsertRequest{Name: "my-db", Tag: "aws/postgres/17/1.0.1", IconURI: "app://qovery-console/terraform"}}
	dispatchStart := time.Date(2026, 9, 24, 15, 0, 0, 0, time.UTC)
	before := dispatchStart.Add(-time.Hour)
	after := dispatchStart.Add(5 * time.Minute)

	t.Run("ignores the deployment in place until the apply one finishes", func(t *testing.T) {
		repo := mocks_test.NewBlueprintRepository(t)
		done := testBlueprint(environmentID, &blueprint.Dispatch{ID: "dispatch-2", Status: blueprint.DispatchStatusRunning, StartedAt: dispatchStart}, &serviceID)
		repo.EXPECT().Update(mock.Anything, blueprintID, request).Return(nil)
		repo.EXPECT().Deploy(mock.Anything, blueprintID).Return("dispatch-2", nil)
		repo.EXPECT().Get(mock.Anything, blueprintID).Return(done, nil).Once()
		repo.EXPECT().GetServiceStatus(mock.Anything, environmentID.String(), blueprint.ServiceTypeTerraform, serviceID).Return(&blueprint.ServiceStatus{State: "DEPLOYED", LastDeploymentDate: &before}, nil).Once()
		repo.EXPECT().GetServiceStatus(mock.Anything, environmentID.String(), blueprint.ServiceTypeTerraform, serviceID).Return(&blueprint.ServiceStatus{State: "DEPLOYED", LastDeploymentDate: &after}, nil).Once()

		bp, err := newTestBlueprintService(repo).Update(context.Background(), blueprintID, request)
		require.NoError(t, err)
		assert.Equal(t, done, bp)
	})

	t.Run("patch error stops before deploying", func(t *testing.T) {
		repo := mocks_test.NewBlueprintRepository(t)
		repo.EXPECT().Update(mock.Anything, blueprintID, request).Return(errors.New("boom"))

		bp, err := newTestBlueprintService(repo).Update(context.Background(), blueprintID, request)
		assert.Nil(t, bp)
		assert.ErrorContains(t, err, blueprint.ErrFailedToUpdateBlueprint.Error())
	})
}

func TestBlueprintServiceDelete(t *testing.T) {
	t.Parallel()
	environmentID := uuid.New()
	blueprintID := uuid.NewString()
	serviceID := uuid.NewString()
	notFound := apierrors.NewAPIError(apierrors.APIActionRead, apierrors.APIResourceBlueprint, blueprintID, &http.Response{StatusCode: http.StatusNotFound}, errors.New("not found"))

	t.Run("already gone", func(t *testing.T) {
		repo := mocks_test.NewBlueprintRepository(t)
		repo.EXPECT().Get(mock.Anything, blueprintID).Return(nil, notFound)
		require.NoError(t, newTestBlueprintService(repo).Delete(context.Background(), blueprintID))
	})

	t.Run("no service to delete", func(t *testing.T) {
		repo := mocks_test.NewBlueprintRepository(t)
		repo.EXPECT().Get(mock.Anything, blueprintID).Return(testBlueprint(environmentID, nil, nil), nil)
		require.NoError(t, newTestBlueprintService(repo).Delete(context.Background(), blueprintID))
	})

	t.Run("deletes the service and waits for the blueprint to go", func(t *testing.T) {
		repo := mocks_test.NewBlueprintRepository(t)
		bp := testBlueprint(environmentID, nil, &serviceID)
		repo.EXPECT().Get(mock.Anything, blueprintID).Return(bp, nil).Once()
		repo.EXPECT().DeleteService(mock.Anything, blueprint.ServiceTypeTerraform, serviceID).Return(nil)
		repo.EXPECT().Get(mock.Anything, blueprintID).Return(bp, nil).Once()
		repo.EXPECT().Get(mock.Anything, blueprintID).Return(nil, notFound).Once()
		require.NoError(t, newTestBlueprintService(repo).Delete(context.Background(), blueprintID))
	})

	t.Run("service delete error", func(t *testing.T) {
		repo := mocks_test.NewBlueprintRepository(t)
		repo.EXPECT().Get(mock.Anything, blueprintID).Return(testBlueprint(environmentID, nil, &serviceID), nil).Once()
		repo.EXPECT().DeleteService(mock.Anything, blueprint.ServiceTypeTerraform, serviceID).Return(errors.New("boom"))
		err := newTestBlueprintService(repo).Delete(context.Background(), blueprintID)
		assert.ErrorContains(t, err, blueprint.ErrFailedToDeleteBlueprint.Error())
	})
}

func TestBlueprintServiceGet(t *testing.T) {
	t.Parallel()
	environmentID := uuid.New()
	blueprintID := uuid.NewString()
	serviceID := uuid.NewString()

	repo := mocks_test.NewBlueprintRepository(t)
	repo.EXPECT().Get(mock.Anything, blueprintID).Return(testBlueprint(environmentID, nil, &serviceID), nil)
	repo.EXPECT().GetServiceStatus(mock.Anything, environmentID.String(), blueprint.ServiceTypeTerraform, serviceID).Return(&blueprint.ServiceStatus{State: "DEPLOYMENT_ERROR"}, nil)

	bp, err := newTestBlueprintService(repo).Get(context.Background(), blueprintID)
	require.NoError(t, err)
	assert.True(t, bp.LastApplyFailed())
}

func TestBlueprintServiceResolveLatestTag(t *testing.T) {
	t.Parallel()
	environmentID := uuid.NewString()
	organizationID := uuid.NewString()
	catalog := []blueprint.CatalogEntry{
		{CatalogVersion: blueprint.CatalogVersion{Provider: "AWS", ServiceFamily: "postgres", ServiceVersion: "16"}, LatestTag: "AWS/postgres/16/4.1.0"},
		{CatalogVersion: blueprint.CatalogVersion{Provider: "AWS", ServiceFamily: "postgres", ServiceVersion: "17"}, LatestTag: "AWS/postgres/17/4.1.0"},
	}
	newRepo := func(t *testing.T) *mocks_test.BlueprintRepository {
		repo := mocks_test.NewBlueprintRepository(t)
		repo.EXPECT().GetOrganizationID(mock.Anything, environmentID).Return(organizationID, nil)
		repo.EXPECT().ListCatalog(mock.Anything, organizationID).Return(catalog, nil)
		return repo
	}

	t.Run("latest tag of the major version", func(t *testing.T) {
		tag, err := newTestBlueprintService(newRepo(t)).ResolveLatestTag(context.Background(), environmentID, blueprint.CatalogVersion{Provider: "AWS", ServiceFamily: "postgres", ServiceVersion: "17"})
		require.NoError(t, err)
		assert.Equal(t, "AWS/postgres/17/4.1.0", tag)
	})

	t.Run("unknown major lists the available ones", func(t *testing.T) {
		_, err := newTestBlueprintService(newRepo(t)).ResolveLatestTag(context.Background(), environmentID, blueprint.CatalogVersion{Provider: "AWS", ServiceFamily: "postgres", ServiceVersion: "12"})
		assert.ErrorIs(t, err, blueprint.ErrCatalogVersionNotFound)
		assert.ErrorContains(t, err, "available: AWS/postgres/16, AWS/postgres/17")
	})

	t.Run("unknown service", func(t *testing.T) {
		_, err := newTestBlueprintService(newRepo(t)).ResolveLatestTag(context.Background(), environmentID, blueprint.CatalogVersion{Provider: "AWS", ServiceFamily: "oracle", ServiceVersion: "19"})
		assert.ErrorIs(t, err, blueprint.ErrCatalogVersionNotFound)
		assert.NotContains(t, err.Error(), "available")
	})
}
