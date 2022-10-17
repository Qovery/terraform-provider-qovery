package storage

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Storages []Storage

var (
	// ErrInvalidStorage is the error return if a Storage is invalid.
	ErrInvalidStorage = errors.New("invalid storage")
	// ErrInvalidStorages is the error return if a Storages is invalid.
	ErrInvalidStorages = errors.New("invalid storages")
	// ErrInvalidStorageIDParam is returned if the storage id param is invalid.
	ErrInvalidStorageIDParam = errors.New("invalid storage id param")
	// ErrInvalidSizeParam is returned if the size param is invalid.
	ErrInvalidSizeParam = errors.New("invalid size param")
	// ErrInvalidMountPointParam is returned if the mount point param is invalid.
	ErrInvalidMountPointParam = errors.New("invalid mount point param")
	// ErrInvalidTypeParam is returned if the type param is invalid.
	ErrInvalidTypeParam = errors.New("invalid type param")
	// ErrInvalidUpsertRequest is returned if the create request is invalid.
	ErrInvalidUpsertRequest = errors.New("invalid storage upsert request")
)

// Validate returns an error to tell whether the Storages' domain model is valid or not.
func (ss Storages) Validate() error {
	for _, it := range ss {
		if err := it.Validate(); err != nil {
			return errors.Wrap(err, ErrInvalidStorages.Error())
		}
	}

	return nil
}

// IsValid returns a bool to tell whether the Storages' domain model is valid or not.
func (ss Storages) IsValid() bool {
	return ss.Validate() == nil
}

type Storage struct {
	ID         uuid.UUID
	Type       Type   `validate:"required"`
	Size       int32  `validate:"required"`
	MountPoint string `validate:"required"`
}

// Validate returns an error to tell whether the Storage domain model is valid or not.
func (s Storage) Validate() error {
	if err := validator.New().Struct(s); err != nil {
		return errors.Wrap(err, ErrInvalidStorage.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the Storage domain model is valid or not.
func (s Storage) IsValid() bool {
	return s.Validate() == nil
}

// NewStorageParams represents the arguments needed to create a Storage.
type NewStorageParams struct {
	StorageID  string
	Type       string
	Size       int32
	MountPoint string
}

// NewStorage returns a new instance of a Storage domain model.
func NewStorage(params NewStorageParams) (*Storage, error) {
	storageUUID, err := uuid.Parse(params.StorageID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidStorageIDParam.Error())
	}

	storageType, err := NewTypeFromString(params.Type)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidTypeParam.Error())
	}

	if params.Size < 0 {
		return nil, ErrInvalidSizeParam
	}

	if params.MountPoint == "" {
		return nil, ErrInvalidMountPointParam
	}

	v := &Storage{
		ID:         storageUUID,
		Type:       *storageType,
		Size:       params.Size,
		MountPoint: params.MountPoint,
	}

	if err := v.Validate(); err != nil {
		return nil, err
	}

	return v, nil
}

// UpsertRequest represents the parameters needed to create & update a Variable.
type UpsertRequest struct {
	Type       string `validate:"required"`
	Size       int32  `validate:"required"`
	MountPoint string `validate:"required"`

	ID *string
}

// Validate returns an error to tell whether the UpsertRequest is valid or not.
func (r UpsertRequest) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertRequest.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the UpsertRequest is valid or not.
func (r UpsertRequest) IsValid() bool {
	return r.Validate() == nil
}
