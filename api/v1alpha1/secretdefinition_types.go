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

// DataSource represents the actual source of truth path for a secret
type DataSource struct {
	// Path to the actual secret
	Path string `json:"path"`
	// Key where the actual secret is stored
	Key string `json:"key"`
	// Encoding type for the secret. Only base64 supported. Optional
	Encoding string `json:"encoding,omitempty"`
}

// SecretDefinitionSpec defines the desired state of SecretDefinition
type SecretDefinitionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Name    string                `json:"name"`
	Type    string                `json:"type,omitempty"`
	KeysMap map[string]DataSource `json:"keysMap"`
}

// SecretDefinitionStatus defines the observed state of SecretDefinition
type SecretDefinitionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// SecretDefinition is the Schema for the secretdefinitions API
type SecretDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretDefinitionSpec   `json:"spec,omitempty"`
	Status SecretDefinitionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecretDefinitionList contains a list of SecretDefinition
type SecretDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretDefinition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecretDefinition{}, &SecretDefinitionList{})
}
