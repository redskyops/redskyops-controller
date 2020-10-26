/*
Copyright 2020 GramLabs, Inc.

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

package experiment

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO This whole thing should move over to redskyops-go

// Application represents a description of an application to run experiments on.
// +kubebuilder:object:generate=true
// +kubebuilder:object:root=true
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Resources are references to application resources to consider in the generation of the experiment.
	// These strings are the same format as used by Kustomize.
	Resources []string `json:"resources,omitempty"`

	// Cost is used to identify which parts of the application impact the cost of running the application.
	Cost *Cost `json:"cost,omitempty"`

	// CloudProvider is used to provide details about the hosting environment the application is run in.
	CloudProvider *CloudProvider `json:"cloudProvider,omitempty"`

	// Parameters specifies additional details about the experiment parameters.
	Parameters *Parameters `json:"parameters,omitempty"`

	// TODO We should have a qualityOfService: section were you can specify things like
	// the percentage of the max that resources are expected to use (then we add both limits and requests and a constraint)
	// or the "max latency" for the application (we could add a non-optimized metric to capture it).
}

// +kubebuilder:object:generate=true
type Cost struct {
	// Labels of the pods which should be considered when collecting cost information.
	Labels map[string]string `json:"labels,omitempty"`
}

// +kubebuilder:object:generate=true
type CloudProvider struct {
	// Generic cloud provider configuration.
	*GenericCloudProvider `json:",inline"`
	// Configuration specific to Google Cloud Platform.
	GCP *GoogleCloudPlatform `json:"gcp,omitempty"`
	// Configuration specific to Amazon Web Services.
	AWS *AmazonWebServices `json:"aws,omitempty"`
}

// +kubebuilder:object:generate=true
type GoogleCloudPlatform struct {
	// Per-resource cost weightings.
	Cost corev1.ResourceList `json:"cost,omitempty"`
}

// +kubebuilder:object:generate=true
type AmazonWebServices struct {
	// Per-resource cost weightings.
	Cost corev1.ResourceList `json:"cost,omitempty"`
}

// +kubebuilder:object:generate=true
type GenericCloudProvider struct {
	// Per-resource cost weightings.
	Cost corev1.ResourceList `json:"cost,omitempty"`
}

// +kubebuilder:object:generate=true
type Parameters struct {
	// Information related to the discovery of container resources parameters like CPU and memory.
	ContainerResources *ContainerResources `json:"containerResources,omitempty"`
}

// +kubebuilder:object:generate=true
type ContainerResources struct {
	// Labels of Kubernetes objects to consider when generating container resources patches.
	Labels map[string]string `json:"labels,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Application{})
}