package image_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/image"
)

func TestImageValidate(t *testing.T) {
	// setup:
	testCases := []struct {
		description   string
		name          string
		tag           string
		expectedError error
	}{
		{description: "case 1: name and tag are blank", name: "", tag: "", expectedError: image.ErrInvalidNameParam},
		{description: "case 2: name is set and tag is blank", name: "set", tag: "", expectedError: image.ErrInvalidTagParam},
		{description: "case 3: name is blank and tag is set", name: "", tag: "set", expectedError: image.ErrInvalidNameParam},
		{description: "case 4: name and tag are set", name: "set", tag: "set", expectedError: nil},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// execute:
			i := image.Image{
				RegistryID: uuid.New(),
				Name:       tc.name,
				Tag:        tc.tag,
			}

			// verify:
			assert.Equal(t, tc.expectedError, i.Validate())
		})
	}
}

func TestNewImage(t *testing.T) {
	// setup:
	validRegistryID := uuid.New()
	testCases := []struct {
		description    string
		params         image.NewImageParams
		expectedResult *image.Image
		expectedError  error
	}{
		{
			description: "case 1: all params blanks",
			params: image.NewImageParams{
				RegistryID: "",
				Name:       "",
				Tag:        "",
			},
			expectedError:  errors.Wrap(errors.New("invalid UUID length: 0"), image.ErrInvalidRegistryIDParam.Error()),
			expectedResult: nil,
		},
		{
			description: "case 2: invalid registry ID",
			params: image.NewImageParams{
				RegistryID: "toto",
				Name:       "",
				Tag:        "",
			},
			expectedError:  errors.Wrap(errors.New("invalid UUID length: 4"), image.ErrInvalidRegistryIDParam.Error()),
			expectedResult: nil,
		},
		{
			description: "case 3: name is blank",
			params: image.NewImageParams{
				RegistryID: uuid.New().String(),
				Name:       "",
				Tag:        "main",
			},
			expectedError:  image.ErrInvalidNameParam,
			expectedResult: nil,
		},
		{
			description: "case 4: tag is blank",
			params: image.NewImageParams{
				RegistryID: uuid.New().String(),
				Name:       "toto",
				Tag:        "",
			},
			expectedError:  image.ErrInvalidTagParam,
			expectedResult: nil,
		},
		{
			description: "case 5: all properly set",
			params: image.NewImageParams{
				RegistryID: validRegistryID.String(),
				Name:       "toto",
				Tag:        "main",
			},
			expectedError: nil,
			expectedResult: &image.Image{
				RegistryID: validRegistryID,
				Name:       "toto",
				Tag:        "main",
			},
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// execute:
			i, err := image.NewImage(tc.params)

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
