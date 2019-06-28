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
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TestSuiteConditionType string
type Status string
type TestStatus string

const (
	StatusTrue    Status = "True"
	StatusFalse   Status = "False"
	StatusUnknown Status = "Unknown"

	// SuiteUninitialized is when suite has not yet determined tests to run
	// TODO(aszecowka) set it as a default values/initialized value (https://github.com/kyma-incubator/octopus/issues/11)
	SuiteUninitialized TestSuiteConditionType = "Uninitialized"
	// When tests are running
	SuiteRunning TestSuiteConditionType = "Running"
	// When suite is finished and there were configuration problems, like missing test image
	SuiteError TestSuiteConditionType = "Error"
	// When suite is finished and there were failing tests
	SuiteFailed TestSuiteConditionType = "Failed"
	// When all tests passed
	SuiteSucceeded TestSuiteConditionType = "Succeeded"

	// TestStatus represents status of a given test (test-kubeless) , not a test execution, because we can have many
	// executions of the same tests (in case of MaxRetries>0 or Count>1)
	//
	// Test is not yet scheduled
	TestNotYetScheduled TestStatus = "NotYetScheduled"
	// Test is running
	TestScheduled TestStatus = "Scheduled" // TODO(aszecowka)(later) do we need both of them?
	TestRunning   TestStatus = "Running"
	TestUnknown   TestStatus = "Unknown"
	TestFailed    TestStatus = "Failed"
	TestSucceeded TestStatus = "Succeeded"
	TestSkipped   TestStatus = "Skipped"

	ReasonErrorOnInitialization = "initializationFailure"
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
	// Find test definitions by its labels.
	// TestDefinition should match AT LEAST one expression listed here to be executed.
	// For a complete grammar see: https://github.com/kubernetes/apimachinery/blob/master/pkg/labels/selector.go#L811
	MatchLabelExpressions []string `json:"matchLabels,omitempty"`
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
	Reason  string                 `json:"reason,omitempty"`
	Message string                 `json:"message,omitempty"`
}

// TestResult gathers all executions for given TestDefinition
type TestResult struct {
	// Test name
	Name                string          `json:"name"`
	Namespace           string          `json:"namespace"`
	Status              TestStatus      `json:"status"`
	Executions          []TestExecution `json:"executions"`
	DisabledConcurrency bool            `json:"disabledConcurrency,omitempty"`
}

// TestExecution provides status for given test execution
type TestExecution struct {
	// ID is equivalent to a testing Pod name
	ID             string       `json:"id"`
	PodPhase       v1.PodPhase  `json:"podPhase"`
	StartTime      *metav1.Time `json:"startTime,inline,omitempty"`
	CompletionTime *metav1.Time `json:"completionTime,inline,omitempty"`
	Reason         string       `json:"reason,omitempty"`
	Message        string       `json:"message,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ClusterTestSuite{}, &ClusterTestSuiteList{})
}
