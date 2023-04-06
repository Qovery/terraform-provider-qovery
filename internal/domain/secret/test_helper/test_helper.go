package test_helper

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

var (
	DefaultValidSecret = secret.Secret{
		ID:    uuid.New(),
		Scope: variable.ScopeApplication,
		Key:   "SecretKey",
	}

	DefaultValidSecretParams = secret.NewSecretParams{
		SecretID: uuid.New().String(),
		Scope:    variable.ScopeApplication.String(),
		Key:      "SecretKey",
	}

	DefaultInvalidSecret = secret.Secret{
		ID:    uuid.New(),
		Scope: variable.ScopeApplication,
		Key:   "",
	}

	DefaultInvalidSecretParams = secret.NewSecretParams{
		SecretID: uuid.New().String(),
		Scope:    variable.ScopeApplication.String(),
		Key:      "",
	}

	DefaultInvalidSecretParamsError = errors.New("invalid secret: Key: 'Secret.Key' Error:Field validation for 'Key' failed on the 'required' tag")
)
