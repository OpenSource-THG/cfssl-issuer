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

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CfsslIssuer is the Schema for the cfsslissuers API
type CfsslIssuer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CfsslIssuerSpec   `json:"spec,omitempty"`
	Status CfsslIssuerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CfsslIssuerList contains a list of CfsslIssuer
type CfsslIssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CfsslIssuer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CfsslIssuer{}, &CfsslIssuerList{}, &CfsslClusterIssuer{}, &CfsslClusterIssuerList{})
}

func (ci *CfsslIssuer) IsReady() bool {
	if ci == nil {
		return false
	}

	for _, cond := range ci.Status.Conditions {
		if cond.Type == ConditionReady && cond.Status == ConditionTrue {
			return true
		}
	}

	return false
}
