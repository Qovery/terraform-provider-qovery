// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks_test

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	status "github.com/qovery/terraform-provider-qovery/internal/domain/status"
)

// DeploymentRepository is an autogenerated mock type for the Repository type
type DeploymentRepository struct {
	mock.Mock
}

type DeploymentRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *DeploymentRepository) EXPECT() *DeploymentRepository_Expecter {
	return &DeploymentRepository_Expecter{mock: &_m.Mock}
}

// Deploy provides a mock function with given fields: ctx, resourceID, version
func (_m *DeploymentRepository) Deploy(ctx context.Context, resourceID string, version string) (*status.Status, error) {
	ret := _m.Called(ctx, resourceID, version)

	var r0 *status.Status
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *status.Status); ok {
		r0 = rf(ctx, resourceID, version)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*status.Status)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, resourceID, version)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeploymentRepository_Deploy_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Deploy'
type DeploymentRepository_Deploy_Call struct {
	*mock.Call
}

// Deploy is a helper method to define mock.On call
//   - ctx context.Context
//   - resourceID string
//   - version string
func (_e *DeploymentRepository_Expecter) Deploy(ctx interface{}, resourceID interface{}, version interface{}) *DeploymentRepository_Deploy_Call {
	return &DeploymentRepository_Deploy_Call{Call: _e.mock.On("Deploy", ctx, resourceID, version)}
}

func (_c *DeploymentRepository_Deploy_Call) Run(run func(ctx context.Context, resourceID string, version string)) *DeploymentRepository_Deploy_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *DeploymentRepository_Deploy_Call) Return(_a0 *status.Status, _a1 error) *DeploymentRepository_Deploy_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetStatus provides a mock function with given fields: ctx, resourceID
func (_m *DeploymentRepository) GetStatus(ctx context.Context, resourceID string) (*status.Status, error) {
	ret := _m.Called(ctx, resourceID)

	var r0 *status.Status
	if rf, ok := ret.Get(0).(func(context.Context, string) *status.Status); ok {
		r0 = rf(ctx, resourceID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*status.Status)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, resourceID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeploymentRepository_GetStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetStatus'
type DeploymentRepository_GetStatus_Call struct {
	*mock.Call
}

// GetStatus is a helper method to define mock.On call
//   - ctx context.Context
//   - resourceID string
func (_e *DeploymentRepository_Expecter) GetStatus(ctx interface{}, resourceID interface{}) *DeploymentRepository_GetStatus_Call {
	return &DeploymentRepository_GetStatus_Call{Call: _e.mock.On("GetStatus", ctx, resourceID)}
}

func (_c *DeploymentRepository_GetStatus_Call) Run(run func(ctx context.Context, resourceID string)) *DeploymentRepository_GetStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *DeploymentRepository_GetStatus_Call) Return(_a0 *status.Status, _a1 error) *DeploymentRepository_GetStatus_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Restart provides a mock function with given fields: ctx, resourceID
func (_m *DeploymentRepository) Restart(ctx context.Context, resourceID string) (*status.Status, error) {
	ret := _m.Called(ctx, resourceID)

	var r0 *status.Status
	if rf, ok := ret.Get(0).(func(context.Context, string) *status.Status); ok {
		r0 = rf(ctx, resourceID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*status.Status)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, resourceID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeploymentRepository_Restart_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Restart'
type DeploymentRepository_Restart_Call struct {
	*mock.Call
}

// Restart is a helper method to define mock.On call
//   - ctx context.Context
//   - resourceID string
func (_e *DeploymentRepository_Expecter) Restart(ctx interface{}, resourceID interface{}) *DeploymentRepository_Restart_Call {
	return &DeploymentRepository_Restart_Call{Call: _e.mock.On("Restart", ctx, resourceID)}
}

func (_c *DeploymentRepository_Restart_Call) Run(run func(ctx context.Context, resourceID string)) *DeploymentRepository_Restart_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *DeploymentRepository_Restart_Call) Return(_a0 *status.Status, _a1 error) *DeploymentRepository_Restart_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Stop provides a mock function with given fields: ctx, resourceID
func (_m *DeploymentRepository) Stop(ctx context.Context, resourceID string) (*status.Status, error) {
	ret := _m.Called(ctx, resourceID)

	var r0 *status.Status
	if rf, ok := ret.Get(0).(func(context.Context, string) *status.Status); ok {
		r0 = rf(ctx, resourceID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*status.Status)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, resourceID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeploymentRepository_Stop_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Stop'
type DeploymentRepository_Stop_Call struct {
	*mock.Call
}

// Stop is a helper method to define mock.On call
//   - ctx context.Context
//   - resourceID string
func (_e *DeploymentRepository_Expecter) Stop(ctx interface{}, resourceID interface{}) *DeploymentRepository_Stop_Call {
	return &DeploymentRepository_Stop_Call{Call: _e.mock.On("Stop", ctx, resourceID)}
}

func (_c *DeploymentRepository_Stop_Call) Run(run func(ctx context.Context, resourceID string)) *DeploymentRepository_Stop_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *DeploymentRepository_Stop_Call) Return(_a0 *status.Status, _a1 error) *DeploymentRepository_Stop_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

type mockConstructorTestingTNewDeploymentRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewDeploymentRepository creates a new instance of DeploymentRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewDeploymentRepository(t mockConstructorTestingTNewDeploymentRepository) *DeploymentRepository {
	mock := &DeploymentRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
