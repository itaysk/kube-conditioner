/*
Copyright 2019 Itay Shakury.

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

// PodConditionSpec defines the desired state of PodCondition
type PodConditionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	LabelSelector    metav1.LabelSelector `json:"labelSelector"` //TODO: don't allow empty selector, means all, will be confusing
	PrometheusSource *PrometheusSource    `json:"prometheusSource,omitempty"`
	// interval to resample source, in milliseconds
	Interval int `json:"interval"`
}

// PrometheusSource defines the logic behind the condition based on a prometheus source
type PrometheusSource struct {
	// Prometheus server url
	ServerURL string `json:"serverUrl"`
	// Prometheus rule to evaluate. The condition is TRUE when the rule is TRUE.
	Rule string `json:"rule"`
}

// PodConditionStatus defines the observed state of PodCondition
type PodConditionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodCondition is the Schema for the CRD
// +k8s:openapi-gen=true
type PodCondition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PodConditionSpec   `json:"spec,omitempty"`
	Status PodConditionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodConditionList contains a list of PodCondition
type PodConditionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodCondition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PodCondition{}, &PodConditionList{})
}
