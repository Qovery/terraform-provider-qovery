package image

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	// ErrInvalidRegistryIDParam is returned if the registry id param is invalid.
	ErrInvalidRegistryIDParam = errors.New("invalid registry id param")
	// ErrInvalidTagParam is returned if the image tag param is invalid.
	ErrInvalidTagParam = errors.New("invalid tag param")
	// ErrInvalidNameParam is returned if the name param is invalid.
	ErrInvalidNameParam = errors.New("invalid name param")
)

type Image struct {
	RegistryID uuid.UUID `validate:"required"`
	Name       string    `validate:"required"`
	Tag        string    `validate:"required"`
}

func (i Image) Validate() error {
	if i.Name == "" {
		return ErrInvalidNameParam
	}

	if i.Tag == "" {
		return ErrInvalidTagParam
	}

	return nil
}

type NewImageParams struct {
	RegistryID string
	Name       string
	Tag        string
}

func NewImage(params NewImageParams) (*Image, error) {
	registryUUID, err := uuid.Parse(params.RegistryID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidRegistryIDParam.Error())
	}

	image := &Image{
		RegistryID: registryUUID,
		Name:       params.Name,
		Tag:        params.Tag,
	}

	if err := image.Validate(); err != nil {
		return nil, err
	}

	return image, nil
}
