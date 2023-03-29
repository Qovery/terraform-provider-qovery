package test_helper

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

var (
	DefaultValidVariable = variable.Variable{
		ID:    uuid.New(),
		Scope: variable.ScopeApplication,
		Key:   "VariableKey",
		Value: "VariableValue",
	}

	DefaultValidVariableParams = variable.NewVariableParams{
		VariableID: uuid.New().String(),
		Scope:      variable.ScopeApplication.String(),
		Key:        "VariableKey",
		Value:      "VariableValue",
	}

	DefaultInvalidVariable = variable.Variable{
		ID:    uuid.New(),
		Scope: variable.ScopeApplication,
		Key:   "",
		Value: "VariableValue",
	}

	DefaultInvalidVariableParams = variable.NewVariableParams{
		VariableID: uuid.New().String(),
		Scope:      variable.ScopeApplication.String(),
		Key:        "",
		Value:      "VariableValue",
	}

	DefaultInvalidVariableParamsError = errors.New("Key: 'Variable.Key' Error:Field validation for 'Key' failed on the 'required' tag")
)
