/*
Copyright 2024.

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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AssignSpec defines the desired state of Assign.
// It specifies where and how the mutation rules should be applied.
type AssignSpec struct {
	// ApplyTo specifies the target resources (Groups, Kinds, Versions) for mutation.
	ApplyTo []ApplyTo `json:"applyTo"`

	// Location indicates the specific path or field in the resource to mutate.
	Location string `json:"location"`

	// Match defines criteria to match specific Kubernetes resources.
	Match Match `json:"match"`

	// Parameters holds the value assignment logic for mutation.
	Parameters Parameters `json:"parameters"`
}

// ApplyTo specifies the resource types targeted by the mutation rule.
type ApplyTo struct {
	// Groups represent Kubernetes API groups the rule applies to.
	Groups []string `json:"groups"`

	// Kind represents the kinds of Kubernetes resources the rule targets.
	Kind []string `json:"kinds"`

	// Versions represent API versions of the targeted resources.
	Versions []string `json:"versions"`
}

// Match defines criteria for selecting resources to mutate.
type Match struct {
	// Kinds specifies resource kinds to match.
	Kinds []KindSelector `json:"kinds"`

	// NameSpaceSelector selects namespaces based on labels.
	NameSpaceSelector NameSpaceSelector `json:"namespaceSelector"`

	// Scope defines whether the rule applies to cluster-wide or namespaced resources.
	Scope string `json:"scope"`
}

// KindSelector specifies Kubernetes resource kinds and their API groups.
type KindSelector struct {
	// ApiGroups are API groups of the targeted resource.
	ApiGroups []string `json:"apiGroups"`

	// Kinds represent the specific kinds of resources.
	Kinds []string `json:"kinds"`
}

// NameSpaceSelector defines namespace selection based on labels.
type NameSpaceSelector struct {
	// MatchLabels selects namespaces with specific labels.
	MatchLabels map[string]string `json:"matchLabels"`
}

// Parameters holds mutation parameters for the rule.
type Parameters struct {
	// Assign contains the assignment details for the mutation.
	Assign AssignParameters `json:"assign"`
}

// AssignParameters defines the assignment parameters.
type AssignParameters struct {
	// Value holds the dynamic value for mutation.
	// +kubebuilder:validation:XPreserveUnknownFields
	Value apiextensionsv1.JSON `json:"value"`
}

// AssignStatus defines the observed state of Assign.
type AssignStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Assign is the Schema for the assigns API.
type Assign struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AssignSpec   `json:"spec,omitempty"`
	Status AssignStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AssignList contains a list of Assign.
type AssignList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Assign `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Assign{}, &AssignList{})
}
