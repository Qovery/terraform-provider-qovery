package variable_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

func TestNewVariable(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Params        variable.NewVariableParams
		ExpectedError error
	}{
		{
			TestName: "fail_with_invalid_variable_id",
			Params: variable.NewVariableParams{
				Scope: variable.ScopeApplication.String(),
				Key:   gofakeit.Name(),
				Value: gofakeit.Name(),
			},
			ExpectedError: variable.ErrInvalidVariableIDParam,
		},
		{
			TestName: "fail_with_invalid_key",
			Params: variable.NewVariableParams{
				VariableID: gofakeit.UUID(),
				Scope:      variable.ScopeApplication.String(),
				Value:      gofakeit.Name(),
			},
			ExpectedError: variable.ErrInvalidKeyParam,
		},
		{
			TestName: "fail_with_invalid_value",
			Params: variable.NewVariableParams{
				VariableID: gofakeit.UUID(),
				Scope:      variable.ScopeApplication.String(),
				Key:        gofakeit.Name(),
			},
			ExpectedError: variable.ErrInvalidValueParam,
		},
		{
			TestName: "fail_with_invalid_scope",
			Params: variable.NewVariableParams{
				VariableID: gofakeit.UUID(),
				Key:        gofakeit.Name(),
				Value:      gofakeit.Name(),
			},
			ExpectedError: variable.ErrInvalidScopeParam,
		},
		{
			TestName: "success",
			Params: variable.NewVariableParams{
				VariableID: gofakeit.UUID(),
				Scope:      variable.ScopeApplication.String(),
				Key:        gofakeit.Name(),
				Value:      gofakeit.Name(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			v, err := variable.NewVariable(tc.Params)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, v)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, v)
			assert.True(t, v.IsValid())
			assert.Equal(t, tc.Params.VariableID, v.ID.String())
			assert.Equal(t, tc.Params.Key, v.Key)
			assert.Equal(t, tc.Params.Value, v.Value)
		})
	}
}
