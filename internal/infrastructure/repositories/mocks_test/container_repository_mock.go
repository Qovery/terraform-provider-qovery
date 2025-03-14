// Code generated by mockery v2.50.2. DO NOT EDIT.

package mocks_test

import (
	context "context"

	container "github.com/qovery/terraform-provider-qovery/internal/domain/container"

	mock "github.com/stretchr/testify/mock"
)

// ContainerRepository is an autogenerated mock type for the Repository type
type ContainerRepository struct {
	mock.Mock
}

type ContainerRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *ContainerRepository) EXPECT() *ContainerRepository_Expecter {
	return &ContainerRepository_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: ctx, environmentID, request
func (_m *ContainerRepository) Create(ctx context.Context, environmentID string, request container.UpsertRepositoryRequest) (*container.Container, error) {
	ret := _m.Called(ctx, environmentID, request)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 *container.Container
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, container.UpsertRepositoryRequest) (*container.Container, error)); ok {
		return rf(ctx, environmentID, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, container.UpsertRepositoryRequest) *container.Container); ok {
		r0 = rf(ctx, environmentID, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*container.Container)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, container.UpsertRepositoryRequest) error); ok {
		r1 = rf(ctx, environmentID, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ContainerRepository_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type ContainerRepository_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx context.Context
//   - environmentID string
//   - request container.UpsertRepositoryRequest
func (_e *ContainerRepository_Expecter) Create(ctx interface{}, environmentID interface{}, request interface{}) *ContainerRepository_Create_Call {
	return &ContainerRepository_Create_Call{Call: _e.mock.On("Create", ctx, environmentID, request)}
}

func (_c *ContainerRepository_Create_Call) Run(run func(ctx context.Context, environmentID string, request container.UpsertRepositoryRequest)) *ContainerRepository_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(container.UpsertRepositoryRequest))
	})
	return _c
}

func (_c *ContainerRepository_Create_Call) Return(_a0 *container.Container, _a1 error) *ContainerRepository_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ContainerRepository_Create_Call) RunAndReturn(run func(context.Context, string, container.UpsertRepositoryRequest) (*container.Container, error)) *ContainerRepository_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function with given fields: ctx, containerID
func (_m *ContainerRepository) Delete(ctx context.Context, containerID string) error {
	ret := _m.Called(ctx, containerID)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, containerID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ContainerRepository_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type ContainerRepository_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - containerID string
func (_e *ContainerRepository_Expecter) Delete(ctx interface{}, containerID interface{}) *ContainerRepository_Delete_Call {
	return &ContainerRepository_Delete_Call{Call: _e.mock.On("Delete", ctx, containerID)}
}

func (_c *ContainerRepository_Delete_Call) Run(run func(ctx context.Context, containerID string)) *ContainerRepository_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *ContainerRepository_Delete_Call) Return(_a0 error) *ContainerRepository_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ContainerRepository_Delete_Call) RunAndReturn(run func(context.Context, string) error) *ContainerRepository_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: ctx, containerID, advancedSettingsJsonFromState, isTriggeredFromImport
func (_m *ContainerRepository) Get(ctx context.Context, containerID string, advancedSettingsJsonFromState string, isTriggeredFromImport bool) (*container.Container, error) {
	ret := _m.Called(ctx, containerID, advancedSettingsJsonFromState, isTriggeredFromImport)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *container.Container
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, bool) (*container.Container, error)); ok {
		return rf(ctx, containerID, advancedSettingsJsonFromState, isTriggeredFromImport)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, bool) *container.Container); ok {
		r0 = rf(ctx, containerID, advancedSettingsJsonFromState, isTriggeredFromImport)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*container.Container)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, bool) error); ok {
		r1 = rf(ctx, containerID, advancedSettingsJsonFromState, isTriggeredFromImport)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ContainerRepository_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type ContainerRepository_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - containerID string
//   - advancedSettingsJsonFromState string
//   - isTriggeredFromImport bool
func (_e *ContainerRepository_Expecter) Get(ctx interface{}, containerID interface{}, advancedSettingsJsonFromState interface{}, isTriggeredFromImport interface{}) *ContainerRepository_Get_Call {
	return &ContainerRepository_Get_Call{Call: _e.mock.On("Get", ctx, containerID, advancedSettingsJsonFromState, isTriggeredFromImport)}
}

func (_c *ContainerRepository_Get_Call) Run(run func(ctx context.Context, containerID string, advancedSettingsJsonFromState string, isTriggeredFromImport bool)) *ContainerRepository_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(bool))
	})
	return _c
}

func (_c *ContainerRepository_Get_Call) Return(_a0 *container.Container, _a1 error) *ContainerRepository_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ContainerRepository_Get_Call) RunAndReturn(run func(context.Context, string, string, bool) (*container.Container, error)) *ContainerRepository_Get_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function with given fields: ctx, containerID, request
func (_m *ContainerRepository) Update(ctx context.Context, containerID string, request container.UpsertRepositoryRequest) (*container.Container, error) {
	ret := _m.Called(ctx, containerID, request)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 *container.Container
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, container.UpsertRepositoryRequest) (*container.Container, error)); ok {
		return rf(ctx, containerID, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, container.UpsertRepositoryRequest) *container.Container); ok {
		r0 = rf(ctx, containerID, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*container.Container)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, container.UpsertRepositoryRequest) error); ok {
		r1 = rf(ctx, containerID, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ContainerRepository_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type ContainerRepository_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - ctx context.Context
//   - containerID string
//   - request container.UpsertRepositoryRequest
func (_e *ContainerRepository_Expecter) Update(ctx interface{}, containerID interface{}, request interface{}) *ContainerRepository_Update_Call {
	return &ContainerRepository_Update_Call{Call: _e.mock.On("Update", ctx, containerID, request)}
}

func (_c *ContainerRepository_Update_Call) Run(run func(ctx context.Context, containerID string, request container.UpsertRepositoryRequest)) *ContainerRepository_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(container.UpsertRepositoryRequest))
	})
	return _c
}

func (_c *ContainerRepository_Update_Call) Return(_a0 *container.Container, _a1 error) *ContainerRepository_Update_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ContainerRepository_Update_Call) RunAndReturn(run func(context.Context, string, container.UpsertRepositoryRequest) (*container.Container, error)) *ContainerRepository_Update_Call {
	_c.Call.Return(run)
	return _c
}

// NewContainerRepository creates a new instance of ContainerRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewContainerRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *ContainerRepository {
	mock := &ContainerRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
