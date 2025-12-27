package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StaticAPISpec struct {
	Path    string   `json:"path"`
	Methods []Method `json:"methods"`
}

type Method struct {
	Method string `json:"method" yaml:"method"`
	// +kubebuilder:validation:Minimum=100
	// +kubebuilder:validation:Maximum=599
	StatusCode int               `json:"statusCode" yaml:"status-code"`
	Body       string            `json:"body,omitempty" yaml:"body,omitempty"`
	Headers    map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
}

type StaticAPIStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=sapi

type StaticAPI struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StaticAPISpec   `json:"spec,omitempty"`
	Status StaticAPIStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type StaticAPIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StaticAPI `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StaticAPI{}, &StaticAPIList{})
}
