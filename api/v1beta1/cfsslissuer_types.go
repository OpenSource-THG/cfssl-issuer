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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CfsslIssuerSpec defines the desired state of CfsslIssuer
type CfsslIssuerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// URL is an url of a Cfssl Server
	URL string `json:"url"`

	// CABundle is a base64 encoded TLS certificate used to verify connections
	// to the step certificates server. If not set the system root certificates
	// are used to validate the TLS connection.
	CABundle []byte `json:"caBundle"`

	// Profile is signing profile used by the Cfssl Server. If omitted, the
	// default profile will be used
	Profile string `json:"profile,omitempty"`
}

// CfsslIssuerStatus defines the observed state of CfsslIssuer
type CfsslIssuerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	Conditions []CfsslIssuerCondition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="URL",type="string",JSONPath=".spec.url",description="",priority=1
// +kubebuilder:printcolumn:name="Profile",type="string",JSONPath=".spec.profile",description="",priority=1
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=cfsslissuers

// CfsslIssuer is the Schema for the cfsslissuers API
type CfsslIssuer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CfsslIssuerSpec   `json:"spec,omitempty"`
	Status CfsslIssuerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CfsslIssuerList contains a list of CfsslIssuer
type CfsslIssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CfsslIssuer `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status

// CfsslClusterIssuer is the Schema for the cfsslclusterissuers API
type CfsslClusterIssuer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CfsslIssuerSpec   `json:"spec,omitempty"`
	Status CfsslIssuerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CfsslClusterIssuerList contains a list of CfsslIssuer
type CfsslClusterIssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CfsslIssuer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CfsslIssuer{}, &CfsslIssuerList{}, &CfsslClusterIssuer{}, &CfsslClusterIssuerList{})
}

// ConditionType represents a CfsslIssuer condition type.
// +kubebuilder:validation:Enum=Ready
type ConditionType string

const (
	// ConditionReady indicates that a CfsslIssuer is ready for use.
	ConditionReady ConditionType = "Ready"
)

// ConditionStatus represents a condition's status.
// +kubebuilder:validation:Enum=True;False;Unknown
type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in
// the condition; "ConditionFalse" means a resource is not in the condition;
// "ConditionUnknown" means kubernetes can't decide if a resource is in the
// condition or not. In the future, we could add other intermediate
// conditions, e.g. ConditionDegraded.
const (
	// ConditionTrue represents the fact that a given condition is true
	ConditionTrue ConditionStatus = "True"

	// ConditionFalse represents the fact that a given condition is false
	ConditionFalse ConditionStatus = "False"

	// ConditionUnknown represents the fact that a given condition is unknown
	ConditionUnknown ConditionStatus = "Unknown"
)

// CfsslIssuerCondition contains condition information for the cfssl issuer.
type CfsslIssuerCondition struct {
	// Type of the condition, currently ('Ready').
	Type ConditionType `json:"type"`

	// Status of the condition, one of ('True', 'False', 'Unknown').
	// +kubebuilder:validation:Enum=True;False;Unknown
	Status ConditionStatus `json:"status"`

	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	// +optional
	Message string `json:"message,omitempty"`
}
