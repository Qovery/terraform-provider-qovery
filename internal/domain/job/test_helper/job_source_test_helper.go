package test_helper

import (
	"github.com/pkg/errors"
	image_test_helper "github.com/qovery/terraform-provider-qovery/internal/domain/image/test_helper"
	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
)

var (
	DefaultNewJobSourceParams = job.NewJobSourceParams{
		Image:  &image_test_helper.DefaultValidNewImageParams,
		Docker: nil,
	}

	DefaultJobSource = job.Source{
		Image:  &image_test_helper.DefaultValidImage,
		Docker: nil,
	}

	DefaultNewInvalidJobSourceParams = job.NewJobSourceParams{
		Image:  &image_test_helper.DefaultInvalidNewImageParams,
		Docker: nil,
	}

	DefaultInvalidJobSource = job.Source{
		Image:  &image_test_helper.DefaultInvalidImage,
		Docker: nil,
	}

	DefaultInvalidNewJobSourceParamsError = errors.Wrap(image_test_helper.DefaultInvalidNewImageParamsError, job.ErrInvalidJobSourceImageParam.Error())
)
