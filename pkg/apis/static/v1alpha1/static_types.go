package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StaticSpec struct {
	Replicas  *int32                       `json:"replicas,omitempty"`
	Image     string                       `json:"image,omitempty"`
	LogLevel  string                       `json:"logLevel,omitempty"`
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	TLS       *TLSConfig                   `json:"tls,omitempty"`
}

type TLSConfig struct {
	Enabled      bool   `json:"enabled,omitempty"`
	SecretName   string `json:"secretName,omitempty"`
	Certificate  string `json:"certificate,omitempty"`
	Key          string `json:"key,omitempty"`
	CA           string `json:"ca,omitempty"`
	VerifyClient bool   `json:"verifyClient,omitempty"`
}

type StaticStatus struct {
	Ready    bool   `json:"ready,omitempty"`
	Replicas int32  `json:"replicas,omitempty"`
	Message  string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced

type Static struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StaticSpec   `json:"spec,omitempty"`
	Status StaticStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type StaticList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Static `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Static{}, &StaticList{})
}
