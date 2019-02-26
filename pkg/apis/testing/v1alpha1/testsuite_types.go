/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TestSuiteSpec defines the desired state of TestSuite
type TestSuiteSpec struct {
	// How many tests we want to execute at the same time.
	// Depends on cluster size and it's load
	// Default value is 1
	Concurrency int64 `json:"concurrency,omitempty"`
	// You can specify test names explicitly.
	TestNamesSelector []TestDefReference `json:"testNamesSelector,omitempty"`
	// Will run every test that depends of any of the component listed here
	ComponentsSelector []string `json:"componentsSelector,omitempty"`
	// Will run every test if set to true. Default values is false
	AllTestsSelector bool `json:"allTestsSelector,omitempty"`
	// Running all tests from suite cannot take more time that specified here.
	// Default value is 1h
	SuiteTimeout *metav1.Duration `json:"suiteTimeout,inline,omitempty"`
	// If specific TestDefinition does not define timeout, use this one
	// No default value
	DefaultTestTimeout *metav1.Duration `json:"defaultTestTimeout,inline,omitempty"`
	// How many times should I run every test? Default value will be 1.
	Count int64 `json:"count,omitempty"`
	// In case of a failed test, how many times it will be retried.
	// If test failed and on retry it succeeded, Test Suite should be marked as a succeeded.
	// Default value is 0 - no retries.
	// MaxRetries and Count cannot be used mutually.
	MaxRetries int64 `json:"maxRetries,omitempty"`
}

type TestDefReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// TestSuiteStatus defines the observed state of TestSuite
type TestSuiteStatus struct {
	StartTime      *metav1.Time         `json:"startTime,inline,omitempty"`
	CompletionTime *metav1.Time         `json:"completionTime,inline,omitempty"`
	Conditions     []TestSuiteCondition `json:"conditions,omitempty"`
	Results        []TestResult         `json:"results,omitempty"`
}

type TestSuiteConditionType string
type Status string
type TestResultConditionType string

const (
	StatusTrue    Status = "True"
	StatusFalse   Status = "False"
	StatusUnknown Status = "Unknown"

	// SuiteUninitialized is when suite has not yet determined tests to run
	SuiteUninitialized TestSuiteConditionType = "Uninitialized"
	// When tests are running
	SuiteRunning TestSuiteConditionType = "Running"
	// When suite is finished and there were configuration problems, like missing test image
	SuiteError TestSuiteConditionType = "Error"
	// When suite is finished and there were failing tests
	SuiteFailed TestSuiteConditionType = "Failed"
	// When all tests passed
	SuiteSucceed TestSuiteConditionType = "Succeed"

	// Test is not yet scheduled
	TestNotYetScheduled TestResultConditionType = "NotYetScheduled"
	// Test is running
	TestRunning TestResultConditionType = "Running"
	TestError   TestResultConditionType = "Error"
	TestFailed  TestResultConditionType = "Failed"
	TestSucceed TestResultConditionType = "Succeed"
)

type TestSuiteCondition struct {
	Type    TestSuiteConditionType `json:"type"`
	Status  Status                 `json:"status"`
	Reason  string                 `json:"reason"`
	Message string                 `json:"message"`
}

type TestResultCondition struct {
	Type    TestResultConditionType `json:"type"`
	Status  Status                  `json:"status"`
	Reason  string                  `json:"reason"`
	Message string                  `json:"message"`
}

// TestResult for execution of given test.
// If test is retried (maxRetrires > 0), or executed many times (count > 1)
// we will have many test result entries for the same test definition (the same name, namespace) but different ID (testing pod)
type TestResult struct {
	// Test name
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	// Unique ID of specific test execution. Equivalent to testing Pod name
	ID             string                `json:"id"`
	StartTime      *metav1.Time          `json:"startTime,inline,omitempty"`
	CompletionTime *metav1.Time          `json:"completionTime,inline,omitempty"`
	Conditions     []TestResultCondition `json:"conditions,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// TestSuite is the Schema for the testsuites API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type TestSuite struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TestSuiteSpec   `json:"spec,omitempty"`
	Status TestSuiteStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// TestSuiteList contains a list of TestSuite
type TestSuiteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TestSuite `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TestSuite{}, &TestSuiteList{})
}
