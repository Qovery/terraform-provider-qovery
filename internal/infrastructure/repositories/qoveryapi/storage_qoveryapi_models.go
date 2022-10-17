package qoveryapi

import (
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/storage"
)

// newQoveryStoragesRequestFromDomain takes the domain request container.UpsertRequest and turns it into a qovery.ContainerRequest to make the api call.
func newQoveryStoragesRequestFromDomain(requests []storage.UpsertRequest) ([]qovery.ServiceStorageRequestStorageInner, error) {
	storages := make([]qovery.ServiceStorageRequestStorageInner, 0, len(requests))
	for _, r := range requests {
		newStorage, err := newQoveryStorageRequestFromDomain(r)
		if err != nil {
			return nil, err
		}

		storages = append(storages, *newStorage)
	}

	return storages, nil
}

// newQoveryStorageRequestFromDomain takes the domain request storage.UpsertRequest and turns it into a qovery.ServiceStorageRequestStoragesInner to make the api call.
func newQoveryStorageRequestFromDomain(request storage.UpsertRequest) (*qovery.ServiceStorageRequestStorageInner, error) {
	storageType, err := qovery.NewStorageTypeEnumFromValue(request.Type)
	if err != nil {
		return nil, errors.Wrap(err, storage.ErrInvalidUpsertRequest.Error())
	}

	return &qovery.ServiceStorageRequestStorageInner{
		Id:         request.ID,
		Size:       request.Size,
		MountPoint: request.MountPoint,
		Type:       *storageType,
	}, nil

}

func newDomainStoragesFromQovery(qoveryStorages []qovery.ServiceStorageStorageInner) (storage.Storages, error) {
	storages := make(storage.Storages, 0, len(qoveryStorages))
	for _, qoveryStorage := range qoveryStorages {
		newStorage, err := newDomainStorageFromQovery(qoveryStorage)
		if err != nil {
			return nil, err
		}
		storages = append(storages, *newStorage)
	}

	return storages, nil
}

func newDomainStorageFromQovery(qoveryStorage qovery.ServiceStorageStorageInner) (*storage.Storage, error) {
	return storage.NewStorage(storage.NewStorageParams{
		StorageID:  qoveryStorage.Id,
		Type:       string(qoveryStorage.Type),
		Size:       qoveryStorage.Size,
		MountPoint: qoveryStorage.MountPoint,
	})
}
