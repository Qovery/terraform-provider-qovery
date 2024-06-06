package job

import (
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/docker"
	"github.com/qovery/terraform-provider-qovery/internal/domain/image"
)

var (
	ErrInvalidJobSourceImageParam                 = errors.New("invalid image param")
	ErrInvalidJobSourceDockerParam                = errors.New("invalid docker param")
	ErrInvalidJobSourceDockerAndImageAreBothSet   = errors.New("invalid job source: either `Docker` or `Image` should be set, not both")
	ErrInvalidJobSourceNoneOfDockerAndImageAreSet = errors.New("invalid job source: either `Docker` or `Image` should be set")
)

type Source struct {
	Image  *image.Image
	Docker *docker.Docker
}

type SourceResponse struct {
	Image  *qovery.ContainerSource
	Docker *qovery.JobSourceDockerResponse
}

func (s Source) Validate() error {
	if s.Docker == nil && s.Image == nil {
		return ErrInvalidJobSourceNoneOfDockerAndImageAreSet
	}

	if s.Docker != nil && s.Image != nil {
		return ErrInvalidJobSourceDockerAndImageAreBothSet
	}

	if s.Docker != nil {
		if err := s.Docker.Validate(); err != nil {
			return errors.Wrap(err, ErrInvalidJobSourceDockerParam.Error())
		}
	}

	if s.Image != nil {
		if err := s.Image.Validate(); err != nil {
			return errors.Wrap(err, ErrInvalidJobSourceImageParam.Error())
		}
	}

	return nil
}

// Tag returns a string representing job unique tag, will be gather from job source.
func (s Source) Tag() *string {
	if err := s.Validate(); err != nil {
		// Should not happen, condition checked uppon creation
		return nil
	}

	if s.Docker != nil && s.Docker.GitRepository.CommitID != nil {
		return s.Docker.GitRepository.CommitID
	}

	if s.Image != nil {
		return &s.Image.Tag
	}

	return nil
}

// TagOrEmpty returns a string representing job unique tag or empty if doesn't exist.
func (s Source) TagOrEmpty() string {
	if s.Tag() != nil {
		return *s.Tag()
	}

	return ""
}

type NewJobSourceParams struct {
	Image  *image.NewImageParams
	Docker *docker.NewDockerParams
}

func NewJobSource(params NewJobSourceParams) (*Source, error) {
	var err error = nil

	var img *image.Image = nil
	if params.Image != nil {
		img, err = image.NewImage(*params.Image)
		if err != nil {
			return nil, errors.Wrap(err, ErrInvalidJobSourceImageParam.Error())
		}
	}

	var dckr *docker.Docker = nil
	if params.Docker != nil {
		dckr, err = docker.NewDocker(*params.Docker)
		if err != nil {
			return nil, errors.Wrap(err, ErrInvalidJobSourceDockerParam.Error())
		}
	}

	newSource := &Source{
		Image:  img,
		Docker: dckr,
	}

	if err := newSource.Validate(); err != nil {
		return nil, err
	}

	return newSource, nil
}
