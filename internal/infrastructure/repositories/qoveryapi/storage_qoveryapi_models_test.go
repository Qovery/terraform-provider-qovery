package qoveryapi

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/storage"
)

func TestNewDomainStoragesFromQovery(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Storages      []qovery.ServiceStorageStorageInner
		ExpectedError error
	}{
		{
			TestName: "success_with_nil_container",
		},
		{
			TestName: "success",
			Storages: []qovery.ServiceStorageStorageInner{
				{
					Id:         gofakeit.UUID(),
					Size:       10,
					Type:       qovery.STORAGETYPEENUM_FAST_SSD,
					MountPoint: gofakeit.Word(),
				},
				{
					Id:         gofakeit.UUID(),
					Size:       10,
					Type:       qovery.STORAGETYPEENUM_FAST_SSD,
					MountPoint: gofakeit.Word(),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			ss, err := newDomainStoragesFromQovery(tc.Storages)
			assert.NoError(t, err)
			assert.Len(t, tc.Storages, len(ss))

			for idx, s := range ss {
				assert.True(t, s.IsValid())
				assert.Equal(t, tc.Storages[idx].Id, s.ID.String())
				assert.Equal(t, tc.Storages[idx].Size, s.Size)
				assert.Equal(t, tc.Storages[idx].MountPoint, s.MountPoint)
				assert.Equal(t, string(tc.Storages[idx].Type), s.Type.String())
			}
		})
	}
}

func TestNewQoveryStoragesRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Storages      []storage.UpsertRequest
		ExpectedError error
	}{
		{
			TestName: "success_with_nil_container",
		},
		{
			TestName: "success",
			Storages: []storage.UpsertRequest{
				{
					Size:       10,
					Type:       storage.TypeFastSSD.String(),
					MountPoint: gofakeit.Word(),
				},
				{
					ID:         new(gofakeit.UUID()),
					Size:       10,
					Type:       storage.TypeFastSSD.String(),
					MountPoint: gofakeit.Word(),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			ss, err := newQoveryStoragesRequestFromDomain(tc.Storages)
			assert.NoError(t, err)
			assert.Len(t, tc.Storages, len(ss))

			for idx, s := range ss {
				assert.Equal(t, tc.Storages[idx].ID, s.Id)
				assert.Equal(t, tc.Storages[idx].Size, s.Size)
				assert.Equal(t, tc.Storages[idx].MountPoint, s.MountPoint)
				assert.Equal(t, tc.Storages[idx].Type, string(s.Type))
			}
		})
	}
}
