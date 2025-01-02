// Code generated by mockery v2.50.2. DO NOT EDIT.

package mocks_test

import (
	context "context"

	credentials "github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
	mock "github.com/stretchr/testify/mock"
)

// CredentialsAwsRepository is an autogenerated mock type for the AwsRepository type
type CredentialsAwsRepository struct {
	mock.Mock
}

type CredentialsAwsRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *CredentialsAwsRepository) EXPECT() *CredentialsAwsRepository_Expecter {
	return &CredentialsAwsRepository_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: ctx, organizationID, request
func (_m *CredentialsAwsRepository) Create(ctx context.Context, organizationID string, request credentials.UpsertAwsRequest) (*credentials.Credentials, error) {
	ret := _m.Called(ctx, organizationID, request)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 *credentials.Credentials
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, credentials.UpsertAwsRequest) (*credentials.Credentials, error)); ok {
		return rf(ctx, organizationID, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, credentials.UpsertAwsRequest) *credentials.Credentials); ok {
		r0 = rf(ctx, organizationID, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*credentials.Credentials)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, credentials.UpsertAwsRequest) error); ok {
		r1 = rf(ctx, organizationID, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CredentialsAwsRepository_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type CredentialsAwsRepository_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx context.Context
//   - organizationID string
//   - request credentials.UpsertAwsRequest
func (_e *CredentialsAwsRepository_Expecter) Create(ctx interface{}, organizationID interface{}, request interface{}) *CredentialsAwsRepository_Create_Call {
	return &CredentialsAwsRepository_Create_Call{Call: _e.mock.On("Create", ctx, organizationID, request)}
}

func (_c *CredentialsAwsRepository_Create_Call) Run(run func(ctx context.Context, organizationID string, request credentials.UpsertAwsRequest)) *CredentialsAwsRepository_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(credentials.UpsertAwsRequest))
	})
	return _c
}

func (_c *CredentialsAwsRepository_Create_Call) Return(_a0 *credentials.Credentials, _a1 error) *CredentialsAwsRepository_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CredentialsAwsRepository_Create_Call) RunAndReturn(run func(context.Context, string, credentials.UpsertAwsRequest) (*credentials.Credentials, error)) *CredentialsAwsRepository_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function with given fields: ctx, organizationID, credentialsID
func (_m *CredentialsAwsRepository) Delete(ctx context.Context, organizationID string, credentialsID string) error {
	ret := _m.Called(ctx, organizationID, credentialsID)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, organizationID, credentialsID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CredentialsAwsRepository_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type CredentialsAwsRepository_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - organizationID string
//   - credentialsID string
func (_e *CredentialsAwsRepository_Expecter) Delete(ctx interface{}, organizationID interface{}, credentialsID interface{}) *CredentialsAwsRepository_Delete_Call {
	return &CredentialsAwsRepository_Delete_Call{Call: _e.mock.On("Delete", ctx, organizationID, credentialsID)}
}

func (_c *CredentialsAwsRepository_Delete_Call) Run(run func(ctx context.Context, organizationID string, credentialsID string)) *CredentialsAwsRepository_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *CredentialsAwsRepository_Delete_Call) Return(_a0 error) *CredentialsAwsRepository_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *CredentialsAwsRepository_Delete_Call) RunAndReturn(run func(context.Context, string, string) error) *CredentialsAwsRepository_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: ctx, organizationID, credentialsID
func (_m *CredentialsAwsRepository) Get(ctx context.Context, organizationID string, credentialsID string) (*credentials.Credentials, error) {
	ret := _m.Called(ctx, organizationID, credentialsID)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *credentials.Credentials
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*credentials.Credentials, error)); ok {
		return rf(ctx, organizationID, credentialsID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *credentials.Credentials); ok {
		r0 = rf(ctx, organizationID, credentialsID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*credentials.Credentials)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, organizationID, credentialsID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CredentialsAwsRepository_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type CredentialsAwsRepository_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - organizationID string
//   - credentialsID string
func (_e *CredentialsAwsRepository_Expecter) Get(ctx interface{}, organizationID interface{}, credentialsID interface{}) *CredentialsAwsRepository_Get_Call {
	return &CredentialsAwsRepository_Get_Call{Call: _e.mock.On("Get", ctx, organizationID, credentialsID)}
}

func (_c *CredentialsAwsRepository_Get_Call) Run(run func(ctx context.Context, organizationID string, credentialsID string)) *CredentialsAwsRepository_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *CredentialsAwsRepository_Get_Call) Return(_a0 *credentials.Credentials, _a1 error) *CredentialsAwsRepository_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CredentialsAwsRepository_Get_Call) RunAndReturn(run func(context.Context, string, string) (*credentials.Credentials, error)) *CredentialsAwsRepository_Get_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function with given fields: ctx, organizationID, credentialsID, request
func (_m *CredentialsAwsRepository) Update(ctx context.Context, organizationID string, credentialsID string, request credentials.UpsertAwsRequest) (*credentials.Credentials, error) {
	ret := _m.Called(ctx, organizationID, credentialsID, request)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 *credentials.Credentials
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, credentials.UpsertAwsRequest) (*credentials.Credentials, error)); ok {
		return rf(ctx, organizationID, credentialsID, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, credentials.UpsertAwsRequest) *credentials.Credentials); ok {
		r0 = rf(ctx, organizationID, credentialsID, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*credentials.Credentials)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, credentials.UpsertAwsRequest) error); ok {
		r1 = rf(ctx, organizationID, credentialsID, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CredentialsAwsRepository_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type CredentialsAwsRepository_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - ctx context.Context
//   - organizationID string
//   - credentialsID string
//   - request credentials.UpsertAwsRequest
func (_e *CredentialsAwsRepository_Expecter) Update(ctx interface{}, organizationID interface{}, credentialsID interface{}, request interface{}) *CredentialsAwsRepository_Update_Call {
	return &CredentialsAwsRepository_Update_Call{Call: _e.mock.On("Update", ctx, organizationID, credentialsID, request)}
}

func (_c *CredentialsAwsRepository_Update_Call) Run(run func(ctx context.Context, organizationID string, credentialsID string, request credentials.UpsertAwsRequest)) *CredentialsAwsRepository_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(credentials.UpsertAwsRequest))
	})
	return _c
}

func (_c *CredentialsAwsRepository_Update_Call) Return(_a0 *credentials.Credentials, _a1 error) *CredentialsAwsRepository_Update_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CredentialsAwsRepository_Update_Call) RunAndReturn(run func(context.Context, string, string, credentials.UpsertAwsRequest) (*credentials.Credentials, error)) *CredentialsAwsRepository_Update_Call {
	_c.Call.Return(run)
	return _c
}

// NewCredentialsAwsRepository creates a new instance of CredentialsAwsRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCredentialsAwsRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *CredentialsAwsRepository {
	mock := &CredentialsAwsRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
