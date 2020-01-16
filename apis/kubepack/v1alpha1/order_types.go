/*
Copyright The Kubepack Authors.

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
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ResourceKindOrder = "Order"
	ResourceOrder     = "order"
	ResourceOrders    = "orders"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=orders,singular=order,categories={kubepack,appscode}
// +kubebuilder:subresource:status
type Order struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              OrderSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            OrderStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type OrderSpec struct {
	Packages []PackageSelection `json:"items" protobuf:"bytes,1,rep,name=items"`
}

type PackageSelection struct {
	Chart *ChartVersionRef `json:"chart,omitempty" protobuf:"bytes,1,opt,name=chart"`
}

type ChartSelection struct {
	ChartRef `json:",inline" protobuf:"bytes,1,opt,name=chartRef"`
	Versions []ChartVersionValues `json:"versions" protobuf:"bytes,2,rep,name=versions"`
}

type ChartVersionValues struct {
	Version string `json:"version" protobuf:"bytes,1,opt,name=version"`
	// +optional
	// +kubebuilder:validation:EmbeddedResource
	// +kubebuilder:pruning:PreserveUnknownFields
	Parameters *runtime.RawExtension `json:"parameters,omitempty" protobuf:"bytes,2,opt,name=parameters"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type OrderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Order `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}

type OrderStatus struct {
	// ObservedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`
}
