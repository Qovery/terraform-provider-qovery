package test_helper

import (
	"github.com/google/uuid"

	"github.com/qovery/terraform-provider-qovery/internal/domain/image"
)

var (
	DefaultRegistryID = uuid.New()
	DefaultName       = "image-name"
	DefaultTag        = "latest"

	/// Exposed to tests needing to get such object without having to know internal sauce magic
	DefaultValidNewImageParams = image.NewImageParams{
		RegistryID: DefaultRegistryID.String(),
		Name:       DefaultName,
		Tag:        DefaultTag,
	}
	DefaultValidImage = image.Image{
		RegistryID: DefaultRegistryID,
		Name:       DefaultName,
		Tag:        DefaultTag,
	}
	/// Exposed to tests needing to get such object without having to know internal sauce magic
	DefaultInvalidNewImageParams = image.NewImageParams{
		RegistryID: DefaultRegistryID.String(),
		Name:       "",
		Tag:        DefaultTag,
	}
	DefaultInvalidImage = image.Image{
		RegistryID: DefaultRegistryID,
		Name:       "",
		Tag:        DefaultTag,
	}
	DefaultInvalidNewImageParamsError = image.ErrInvalidNameParam
)
