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

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TestDefinition is the Schema for the testdefinitions API
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=testdefinitions,shortName=td
type TestDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec TestDefinitionSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TestDefinitionList contains a list of TestDefinition
type TestDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TestDefinition `json:"items"`
}

// TestDefinitionSpec defines the desired state of TestDefinition
type TestDefinitionSpec struct {
	Template v1.PodTemplateSpec `json:"template"`
	// If there are some problems with given test, we add possibility to don't execute them.
	// On Testsuite level such test should be marked as a skipped.
	// Default value is false
	Skip bool `json:"skip,omitempty"`
	// If test is working on data that can be modified by another test,
	// I would like to run it in separation.
	// Default value is false
	DisableConcurrency bool `json:"disableConcurrency,omitempty"`
	// Test should be interrupted after the timeout.
	// On test suite level such test should be marked as a timeouted.
	// No default value.
	Timeout *metav1.Duration `json:"timeout,inline,omitempty"`
}

func init() {
	SchemeBuilder.Register(&TestDefinition{}, &TestDefinitionList{})
}
