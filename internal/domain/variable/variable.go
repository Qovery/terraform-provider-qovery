package variable

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	// ErrNilVariable is returned if a Variable is nil.
	ErrNilVariable = errors.New("variable cannot be nil")
	// ErrInvalidVariable is the error return if a Variable is invalid.
	ErrInvalidVariable = errors.New("invalid variable")
	// ErrInvalidVariables is the error return if a Variables is invalid.
	ErrInvalidVariables = errors.New("invalid variables")
	// ErrInvalidResourceIDParam is returned if the resource id param is invalid.
	ErrInvalidResourceIDParam = errors.New("invalid resource id param")
	// ErrInvalidVariableIDParam is returned if the variable id param is invalid.
	ErrInvalidVariableIDParam = errors.New("invalid variable id param")
	// ErrInvalidKeyParam is returned if the key param is invalid.
	ErrInvalidKeyParam = errors.New("invalid key param")
	// ErrInvalidValueParam is returned if the value param is invalid.
	ErrInvalidValueParam = errors.New("invalid value param")
	// ErrInvalidScopeParam is returned if the scope param is invalid.
	ErrInvalidScopeParam = errors.New("invalid scope param")
	// ErrInvalidUpsertRequest is returned if the upsert request is invalid.
	ErrInvalidUpsertRequest = errors.New("invalid variable upsert request")
	// ErrInvalidDiffRequest is returned if the diff request is invalid.
	ErrInvalidDiffRequest = errors.New("invalid variable diff request")
)

type Variables []Variable

// Validate returns an error to tell whether the Variables' domain model is valid or not.
func (vv Variables) Validate() error {
	for _, it := range vv {
		if err := it.Validate(); err != nil {
			return errors.Wrap(err, ErrInvalidVariables.Error())
		}
	}

	return nil
}

// IsValid returns a bool to tell whether the Variables domain model is valid or not.
func (vv Variables) IsValid() bool {
	return vv.Validate() == nil
}

type Variable struct {
	ID    uuid.UUID `validate:"required"`
	Scope Scope     `validate:"required"`
	Key   string    `validate:"required"`
	Value string
	Type  string
}

// Validate returns an error to tell whether the Variable domain model is valid or not.
func (v Variable) Validate() error {
	return validator.New().Struct(v)
}

// IsValid returns a bool to tell whether the Variable domain model is valid or not.
func (v Variable) IsValid() bool {
	return v.Validate() == nil
}

// NewVariableParams represents the arguments needed to create a Variable.
type NewVariablesParams = []NewVariableParams
type NewVariableParams struct {
	VariableID string
	Scope      string
	Key        string
	Value      string
	Type       string
}

// NewVariable returns a new instance of a Variable domain model.
func NewVariable(params NewVariableParams) (*Variable, error) {
	variableUUID, err := uuid.Parse(params.VariableID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidVariableIDParam.Error())
	}

	scope, err := NewScopeFromString(params.Scope)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidScopeParam.Error())
	}

	if params.Key == "" {
		return nil, ErrInvalidKeyParam
	}

	v := &Variable{
		ID:    variableUUID,
		Key:   params.Key,
		Value: params.Value,
		Scope: *scope,
		Type:  params.Type,
	}

	if err := v.Validate(); err != nil {
		return nil, errors.Wrap(err, ErrInvalidVariable.Error())
	}

	return v, nil
}

// UpsertRequest represents the parameters needed to create & update a Variable.
type UpsertRequest struct {
	Key   string `validate:"required"`
	Value string
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
