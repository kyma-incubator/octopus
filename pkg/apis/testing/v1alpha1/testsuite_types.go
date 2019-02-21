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
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// How many tests we want to execute at the same time.
	// Depends on cluster size and it's load
	ConcurrencyLevel int64 `json:"concurrency_level,omitempty"`
	// You can specify test names explicitly.
	TestNamesSelector []NamespacedTest `json:"test_name_selector,omitempty"`
	// Will run every test that depends of any of the component listed here
	ComponentsSelector []string `json:"components_selector,omitempty"`
	// Running all tests from suite cannot take more time that specified here
	SuiteTimeout *metav1.Duration `json:"suite_timeout,omitempty"`
	// If specific TestDefinition does not define timeout, use this one
	DefaultTestTimeout *metav1.Duration `json:"default_test_timeout,omitempty"`
	// Should I repeat every test? Default value will be 1
	Repeat int64 `json:"repeat,omitempty"`
	// In case of a failed test, how many times it will be retried.
	// If test failed and on retry it succeeded, Test Suite should be marked as a succeeded.
	MaxRetries int64 `json:"max_retries,omitempty"`
}

type NamespacedTest struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// TestSuiteStatus defines the observed state of TestSuite
type TestSuiteStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	StartTime      metav1.Time          `json:"start_time"`
	CompletionTime *metav1.Time         `json:"completion_time"`
	Conditions     []TestSuiteCondition `json:"conditions,omitempty"`
	Results        []TestResult         `json:"results,omitempty"`
}

type TestSuiteConditionType string
type Status string
type TestResultConditionType string

const (
	StatusTrue    Status = "true"
	StatusFalse   Status = "false"
	StatusUnknown Status = "unknown"

	// SuiteUninitialized is when suite has not yet determined tests to run
	SuiteUninitialized TestSuiteConditionType = "pending"
	// When test are running
	SuiteRunning TestSuiteConditionType = "running"
	// When suite is finished and there were configuration problems, like missing test image
	SuiteError TestSuiteConditionType = "error"
	// When suite is finished and there were failing tests
	SuiteFailed TestSuiteConditionType = "failed"
	// When all tests passed
	SuiteSucceed TestSuiteConditionType = "succeed"

	// Test is not yet scheduled
	TestPending TestResultConditionType = "pending"
	// Test is running
	TestRunning TestResultConditionType = "running"
	TestError   TestResultConditionType = "error"
	TestFailed  TestResultConditionType = "failed"
	TestSucceed TestResultConditionType = "succeed"
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

type TestResult struct {
	// Test name
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	// Unique ID of specific test execution. Equivalent to testing Pod name
	ID         string                `json:"id"`
	Conditions []TestResultCondition `json:"conditions,omitempty"`
	// How many times test was retired
	Retries int64 `json:"retries"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// TestSuite is the Schema for the testsuites API
// +k8s:openapi-gen=true
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
