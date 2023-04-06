package job_test

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/docker"
	docker_test_helper "github.com/qovery/terraform-provider-qovery/internal/domain/docker/test_helper"
	"github.com/qovery/terraform-provider-qovery/internal/domain/image"
	image_test_helper "github.com/qovery/terraform-provider-qovery/internal/domain/image/test_helper"
	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
)

func TestJobSourceValidate(t *testing.T) {
	// setup:
	testCases := []struct {
		description   string
		image         *image.Image
		docker        *docker.Docker
		expectedError error
	}{
		{description: "case 1: image is not valid", image: &image_test_helper.DefaultInvalidImage, docker: nil, expectedError: errors.Wrap(image_test_helper.DefaultInvalidNewImageParamsError, job.ErrInvalidJobSourceImageParam.Error())},
		{description: "case 2: docker is not valid", image: nil, docker: &docker_test_helper.DefaultInvalidDocker, expectedError: errors.Wrap(docker_test_helper.DefaultInvalidNewDockerParamsError, job.ErrInvalidJobSourceDockerParam.Error())},
		{description: "case 3: image is nil", image: nil, docker: &docker_test_helper.DefaultValidDocker, expectedError: nil},
		{description: "case 4: docker is nil", image: &image_test_helper.DefaultValidImage, docker: nil, expectedError: nil},
		{description: "case 5: all fields are set", image: &image_test_helper.DefaultValidImage, docker: &docker_test_helper.DefaultValidDocker, expectedError: job.ErrInvalidJobSourceDockerAndImageAreBothSet},
		{description: "case 6: none fields are set", image: nil, docker: nil, expectedError: job.ErrInvalidJobSourceNoneOfDockerAndImageAreSet},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// execute:
			i := job.JobSource{
				Image:  tc.image,
				Docker: tc.docker,
			}

			// verify:
			if err := i.Validate(); err != nil {
				assert.Equal(t, tc.expectedError.Error(), i.Validate().Error())
			} else {
				assert.Equal(t, tc.expectedError, i.Validate()) // <- should be nil
			}
		})
	}
}

func TestNewJobSource(t *testing.T) {
	// setup:
	testCases := []struct {
		description    string
		params         job.NewJobSourceParams
		expectedResult *job.JobSource
		expectedError  error
	}{
		{
			description: "case 1: invalid docker",
			params: job.NewJobSourceParams{
				Image:  &image_test_helper.DefaultValidNewImageParams,
				Docker: &docker_test_helper.DefaultInvalidNewDockerParams,
			},
			expectedError:  errors.Wrap(docker_test_helper.DefaultInvalidNewDockerParamsError, job.ErrInvalidJobSourceDockerParam.Error()),
			expectedResult: nil,
		},
		{
			description: "case 2: invalid image",
			params: job.NewJobSourceParams{
				Image:  &image_test_helper.DefaultInvalidNewImageParams,
				Docker: &docker_test_helper.DefaultValidNewDockerParams,
			},
			expectedError:  errors.Wrap(image_test_helper.DefaultInvalidNewImageParamsError, job.ErrInvalidJobSourceImageParam.Error()),
			expectedResult: nil,
		},
		{
			description: "case 3: nil image",
			params: job.NewJobSourceParams{
				Image:  nil,
				Docker: &docker_test_helper.DefaultValidNewDockerParams,
			},
			expectedError: nil,
			expectedResult: &job.JobSource{
				Image:  nil,
				Docker: &docker_test_helper.DefaultValidDocker,
			},
		},
		{
			description: "case 4: nil docker",
			params: job.NewJobSourceParams{
				Image:  &image_test_helper.DefaultValidNewImageParams,
				Docker: nil,
			},
			expectedError: nil,
			expectedResult: &job.JobSource{
				Image:  &image_test_helper.DefaultValidImage,
				Docker: nil,
			},
		},
		{
			description: "case 5: nil docker & image",
			params: job.NewJobSourceParams{
				Image:  nil,
				Docker: nil,
			},
			expectedError:  job.ErrInvalidJobSourceNoneOfDockerAndImageAreSet,
			expectedResult: nil,
		},
		{
			description: "case 6: docker & image are both set",
			params: job.NewJobSourceParams{
				Image:  &image_test_helper.DefaultValidNewImageParams,
				Docker: &docker_test_helper.DefaultValidNewDockerParams,
			},
			expectedError:  job.ErrInvalidJobSourceDockerAndImageAreBothSet,
			expectedResult: nil,
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// execute:
			i, err := job.NewJobSource(tc.params)

			// verify:
			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.Equal(t, nil, err)
			}
			assert.Equal(t, tc.expectedResult, i)
		})
	}
}
