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

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// CfsslClusterIssuer is the Schema for the cfsslclusterissuers API
type CfsslClusterIssuer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CfsslIssuerSpec   `json:"spec,omitempty"`
	Status CfsslIssuerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CfsslClusterIssuerList contains a list of CfsslClusterIssuer
type CfsslClusterIssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CfsslClusterIssuer `json:"items"`
}

func (ci *CfsslClusterIssuer) IsReady() bool {
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
