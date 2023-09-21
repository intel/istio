/*
Copyright 2023.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ImdiOperatorSpec defines the desired state of ImdiOperator
type ImdiOperatorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ImdiOperator. Edit imdioperator_types.go to remove/update
	IstioSpec       string                `json:"IstioSpec,omitempty"`
	DevicePlugin    *DevicePluginSetSpec  `json:"DevicePlugin,omitempty"`
	IngressGateways []*IngressGatewaySpec `json:"IngressGateways,omitempty"`
	SetupProxy      ProxySetSpec          `json:"setupproxy,omitempty"`
}

type ProxySetSpec struct {
	Http_proxy  string `json:"http_proxy,omitempty"`
	Https_proxy string `json:"https_proxy,omitempty"`
	No_proxy    string `json:"no_proxy,omitempty"`
}

type DevicePluginSetSpec struct {
	Nfd                    string `json:"nfd,omitempty"`
	Nfd_rules              string `json:"nfd_rules,omitempty"`
	Cert_manager           string `json:"cert_manager,omitempty"`
	Device_plugin_operator string `json:"device_plugin_operator,omitempty"`
}

type IngressGatewaySpec struct {
	Name     string           `json:"name,omitempty"`
	Qat      *QatSetSpec      `json:"Qat,omitempty"`
	CryptoMB *CryptoMBSetSpec `json:"CryptoMB,omitempty"`
	Sgx      *SgxSetSpec      `json:"Sgx,omitempty"`
}

type QatSetSpec struct {
	Instance  int64  `json:"Instance,omitempty"`
	PollDelay string `json:"pollDelay,omitempty"`
}

type CryptoMBSetSpec struct {
	PollDelay string `json:"pollDelay,omitempty"`
}

type SgxSetSpec struct {
	Enable bool `json:"Enable,omitempty"`
}

// ImdiOperatorStatus defines the observed state of ImdiOperator
type ImdiOperatorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +kubebuilder:default="NotSet"
	NfdStatus string `json:"NfdStatus"`
	// +kubebuilder:default="NotSet"
	NfdRuleStatus string `json:"NfdruleStatus"`
	// +kubebuilder:default="NotSet"
	CertManagerStatus string `json:"CertManagerStatus"`
	// +kubebuilder:default="NotSet"
	DevicePluginStatus string `json:"DevicePluginStatus"`
	ProxyEnabled       bool   `json:"proxyenabled"`
	// +kubebuilder:default="NotSet"
	ProxyStatus       string   `json:"proxystatus"`
	// +kubebuilder:default="NotSet"
	QatStatus string `json:"QatStatus"`
	// +kubebuilder:default="NotSet"
	SgxStatus string `json:"SgxStatus"`
	// +kubebuilder:default="NotSet"
	CryptoMBStatus string `json:"CryptoMBStatus"`
	// +kubebuilder:default="NotSet"
	IstioStatus string `json:"IstioStatus"`
	// +kubebuilder:default="NotSet"
	ConfigParseStatus string `json:"ConfigParseStatus"`
	// +kubebuilder:default="NotSet"
	SgxPSWStatus string `json:"SgxPSWStatus"`
	QatDeviceNum int64  `json:"QatDeviceNum"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ImdiOperator is the Schema for the imdioperators API
type ImdiOperator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImdiOperatorSpec   `json:"spec,omitempty"`
	Status ImdiOperatorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ImdiOperatorList contains a list of ImdiOperator
type ImdiOperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ImdiOperator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ImdiOperator{}, &ImdiOperatorList{})
}
