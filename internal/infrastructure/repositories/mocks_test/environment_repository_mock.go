// Code generated by mockery v2.50.2. DO NOT EDIT.

package mocks_test

import (
	context "context"

	environment "github.com/qovery/terraform-provider-qovery/internal/domain/environment"
	mock "github.com/stretchr/testify/mock"
)

// EnvironmentRepository is an autogenerated mock type for the Repository type
type EnvironmentRepository struct {
	mock.Mock
}

type EnvironmentRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *EnvironmentRepository) EXPECT() *EnvironmentRepository_Expecter {
	return &EnvironmentRepository_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: ctx, projectID, request
func (_m *EnvironmentRepository) Create(ctx context.Context, projectID string, request environment.CreateRepositoryRequest) (*environment.Environment, error) {
	ret := _m.Called(ctx, projectID, request)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 *environment.Environment
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, environment.CreateRepositoryRequest) (*environment.Environment, error)); ok {
		return rf(ctx, projectID, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, environment.CreateRepositoryRequest) *environment.Environment); ok {
		r0 = rf(ctx, projectID, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*environment.Environment)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, environment.CreateRepositoryRequest) error); ok {
		r1 = rf(ctx, projectID, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EnvironmentRepository_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type EnvironmentRepository_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx context.Context
//   - projectID string
//   - request environment.CreateRepositoryRequest
func (_e *EnvironmentRepository_Expecter) Create(ctx interface{}, projectID interface{}, request interface{}) *EnvironmentRepository_Create_Call {
	return &EnvironmentRepository_Create_Call{Call: _e.mock.On("Create", ctx, projectID, request)}
}

func (_c *EnvironmentRepository_Create_Call) Run(run func(ctx context.Context, projectID string, request environment.CreateRepositoryRequest)) *EnvironmentRepository_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(environment.CreateRepositoryRequest))
	})
	return _c
}

func (_c *EnvironmentRepository_Create_Call) Return(_a0 *environment.Environment, _a1 error) *EnvironmentRepository_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EnvironmentRepository_Create_Call) RunAndReturn(run func(context.Context, string, environment.CreateRepositoryRequest) (*environment.Environment, error)) *EnvironmentRepository_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function with given fields: ctx, environmentID
func (_m *EnvironmentRepository) Delete(ctx context.Context, environmentID string) error {
	ret := _m.Called(ctx, environmentID)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, environmentID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EnvironmentRepository_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type EnvironmentRepository_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - environmentID string
func (_e *EnvironmentRepository_Expecter) Delete(ctx interface{}, environmentID interface{}) *EnvironmentRepository_Delete_Call {
	return &EnvironmentRepository_Delete_Call{Call: _e.mock.On("Delete", ctx, environmentID)}
}

func (_c *EnvironmentRepository_Delete_Call) Run(run func(ctx context.Context, environmentID string)) *EnvironmentRepository_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *EnvironmentRepository_Delete_Call) Return(_a0 error) *EnvironmentRepository_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *EnvironmentRepository_Delete_Call) RunAndReturn(run func(context.Context, string) error) *EnvironmentRepository_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// Exists provides a mock function with given fields: ctx, environmentId
func (_m *EnvironmentRepository) Exists(ctx context.Context, environmentId string) bool {
	ret := _m.Called(ctx, environmentId)

	if len(ret) == 0 {
		panic("no return value specified for Exists")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string) bool); ok {
		r0 = rf(ctx, environmentId)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// EnvironmentRepository_Exists_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Exists'
type EnvironmentRepository_Exists_Call struct {
	*mock.Call
}

// Exists is a helper method to define mock.On call
//   - ctx context.Context
//   - environmentId string
func (_e *EnvironmentRepository_Expecter) Exists(ctx interface{}, environmentId interface{}) *EnvironmentRepository_Exists_Call {
	return &EnvironmentRepository_Exists_Call{Call: _e.mock.On("Exists", ctx, environmentId)}
}

func (_c *EnvironmentRepository_Exists_Call) Run(run func(ctx context.Context, environmentId string)) *EnvironmentRepository_Exists_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *EnvironmentRepository_Exists_Call) Return(_a0 bool) *EnvironmentRepository_Exists_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *EnvironmentRepository_Exists_Call) RunAndReturn(run func(context.Context, string) bool) *EnvironmentRepository_Exists_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: ctx, environmentID
func (_m *EnvironmentRepository) Get(ctx context.Context, environmentID string) (*environment.Environment, error) {
	ret := _m.Called(ctx, environmentID)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *environment.Environment
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*environment.Environment, error)); ok {
		return rf(ctx, environmentID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *environment.Environment); ok {
		r0 = rf(ctx, environmentID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*environment.Environment)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, environmentID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EnvironmentRepository_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type EnvironmentRepository_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - environmentID string
func (_e *EnvironmentRepository_Expecter) Get(ctx interface{}, environmentID interface{}) *EnvironmentRepository_Get_Call {
	return &EnvironmentRepository_Get_Call{Call: _e.mock.On("Get", ctx, environmentID)}
}

func (_c *EnvironmentRepository_Get_Call) Run(run func(ctx context.Context, environmentID string)) *EnvironmentRepository_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *EnvironmentRepository_Get_Call) Return(_a0 *environment.Environment, _a1 error) *EnvironmentRepository_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EnvironmentRepository_Get_Call) RunAndReturn(run func(context.Context, string) (*environment.Environment, error)) *EnvironmentRepository_Get_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function with given fields: ctx, environmentID, request
func (_m *EnvironmentRepository) Update(ctx context.Context, environmentID string, request environment.UpdateRepositoryRequest) (*environment.Environment, error) {
	ret := _m.Called(ctx, environmentID, request)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 *environment.Environment
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, environment.UpdateRepositoryRequest) (*environment.Environment, error)); ok {
		return rf(ctx, environmentID, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, environment.UpdateRepositoryRequest) *environment.Environment); ok {
		r0 = rf(ctx, environmentID, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*environment.Environment)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, environment.UpdateRepositoryRequest) error); ok {
		r1 = rf(ctx, environmentID, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EnvironmentRepository_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type EnvironmentRepository_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - ctx context.Context
//   - environmentID string
//   - request environment.UpdateRepositoryRequest
func (_e *EnvironmentRepository_Expecter) Update(ctx interface{}, environmentID interface{}, request interface{}) *EnvironmentRepository_Update_Call {
	return &EnvironmentRepository_Update_Call{Call: _e.mock.On("Update", ctx, environmentID, request)}
}

func (_c *EnvironmentRepository_Update_Call) Run(run func(ctx context.Context, environmentID string, request environment.UpdateRepositoryRequest)) *EnvironmentRepository_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(environment.UpdateRepositoryRequest))
	})
	return _c
}

func (_c *EnvironmentRepository_Update_Call) Return(_a0 *environment.Environment, _a1 error) *EnvironmentRepository_Update_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EnvironmentRepository_Update_Call) RunAndReturn(run func(context.Context, string, environment.UpdateRepositoryRequest) (*environment.Environment, error)) *EnvironmentRepository_Update_Call {
	_c.Call.Return(run)
	return _c
}

// NewEnvironmentRepository creates a new instance of EnvironmentRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEnvironmentRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *EnvironmentRepository {
	mock := &EnvironmentRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
