// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import mock "github.com/stretchr/testify/mock"

import v1alpha1 "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"

// StatusProvider is an autogenerated mock type for the StatusProvider type
type StatusProvider struct {
	mock.Mock
}

// GetNextToSchedule provides a mock function with given fields: suite
func (_m *StatusProvider) GetNextToSchedule(suite v1alpha1.ClusterTestSuite) *v1alpha1.TestResult {
	ret := _m.Called(suite)

	var r0 *v1alpha1.TestResult
	if rf, ok := ret.Get(0).(func(v1alpha1.ClusterTestSuite) *v1alpha1.TestResult); ok {
		r0 = rf(suite)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.TestResult)
		}
	}

	return r0
}

// MarkAsScheduled provides a mock function with given fields: status, testName, testNs, podName
func (_m *StatusProvider) MarkAsScheduled(status v1alpha1.TestSuiteStatus, testName string, testNs string, podName string) (v1alpha1.TestSuiteStatus, error) {
	ret := _m.Called(status, testName, testNs, podName)

	var r0 v1alpha1.TestSuiteStatus
	if rf, ok := ret.Get(0).(func(v1alpha1.TestSuiteStatus, string, string, string) v1alpha1.TestSuiteStatus); ok {
		r0 = rf(status, testName, testNs, podName)
	} else {
		r0 = ret.Get(0).(v1alpha1.TestSuiteStatus)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(v1alpha1.TestSuiteStatus, string, string, string) error); ok {
		r1 = rf(status, testName, testNs, podName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
