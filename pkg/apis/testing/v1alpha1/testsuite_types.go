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

type TestSuiteConditionType string
type Status string
type TestExecutionStatus string

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
	TestNotYetScheduled TestExecutionStatus = "NotYetScheduled"
	// Test is running
	TestRunning TestExecutionStatus = "Running"
	TestError   TestExecutionStatus = "Error"
	TestFailed  TestExecutionStatus = "Failed"
	TestSucceed TestExecutionStatus = "Succeed"
	TestSkipped TestExecutionStatus = "Skipped"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// ClusterTestSuite is the Schema for the testsuites API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=clustertestsuites,shortName=cts
type ClusterTestSuite struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TestSuiteSpec   `json:"spec,omitempty"`
	Status TestSuiteStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// ClusterTestSuiteList contains a list of ClusterTestSuite
type ClusterTestSuiteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterTestSuite `json:"items"`
}

// TestSuiteSpec defines the desired state of ClusterTestSuite
type TestSuiteSpec struct {
	// How many tests we want to execute at the same time.
	// Depends on cluster size and it's load.
	// Default value is 1
	Concurrency int64 `json:"concurrency,omitempty"`
	// Decide which tests to execute. If not provided execute all tests
	Selectors TestsSelector `json:"selectors,omitempty"`
	// Running all tests from suite cannot take more time that specified here.
	// Default value is 1h
	SuiteTimeout *metav1.Duration `json:"suiteTimeout,inline,omitempty"`
	// How many times should I run every test? Default value is 1.
	Count int64 `json:"count,omitempty"`
	// In case of a failed test, how many times it will be retried.
	// If test failed and on retry it succeeded, Test Suite should be marked as a succeeded.
	// Default value is 0 - no retries.
	// MaxRetries and Count cannot be used mutually.
	MaxRetries int64 `json:"maxRetries,omitempty"`
}

type TestsSelector struct {
	// Find test definitions by it's name
	MatchNames []TestDefReference `json:"matchNames,omitempty"`
	// Find test definitions by it's labels.
	// TestDefinition should have AT LEAST one label listed here to be executed.
	// Label value is irrelevant.
	MatchLabels []string `json:"matchLabels,omitempty"`
}

type TestDefReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// TestSuiteStatus defines the observed state of ClusterTestSuite
type TestSuiteStatus struct {
	StartTime      *metav1.Time         `json:"startTime,inline,omitempty"`
	CompletionTime *metav1.Time         `json:"completionTime,inline,omitempty"`
	Conditions     []TestSuiteCondition `json:"conditions,omitempty"`
	Results        []TestResult         `json:"results,omitempty"`
}

type TestSuiteCondition struct {
	Type    TestSuiteConditionType `json:"type"`
	Status  Status                 `json:"status"`
	Reason  string                 `json:"reason"`
	Message string                 `json:"message"`
}

// TestResult gathers all executions for given TestDefinition
type TestResult struct {
	// Test name
	Name       string          `json:"name"`
	Namespace  string          `json:"namespace"`
	Executions []TestExecution `json:"executions"`
}

// TestExecution provides status for given test execution
type TestExecution struct {
	ID             string              `json:"id"` // ID is equivalent to a testing Pod name
	Status         TestExecutionStatus `json:"status"`
	StartTime      *metav1.Time        `json:"startTime,inline,omitempty"`
	CompletionTime *metav1.Time        `json:"completionTime,inline,omitempty"`
	Reason         string              `json:"reason"`
	Message        string              `json:"message"`
}

func init() {
	SchemeBuilder.Register(&ClusterTestSuite{}, &ClusterTestSuiteList{})
}
