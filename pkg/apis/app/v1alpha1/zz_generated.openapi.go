// +build !ignore_autogenerated

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1alpha1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"./pkg/apis/app/v1alpha1.AkkaCluster":       schema_pkg_apis_app_v1alpha1_AkkaCluster(ref),
		"./pkg/apis/app/v1alpha1.AkkaClusterSpec":   schema_pkg_apis_app_v1alpha1_AkkaClusterSpec(ref),
		"./pkg/apis/app/v1alpha1.AkkaClusterStatus": schema_pkg_apis_app_v1alpha1_AkkaClusterStatus(ref),
	}
}

func schema_pkg_apis_app_v1alpha1_AkkaCluster(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "AkkaCluster is the Schema for the akkaclusters API",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/api/apps/v1.DeploymentSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/app/v1alpha1.AkkaClusterStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"./pkg/apis/app/v1alpha1.AkkaClusterStatus", "k8s.io/api/apps/v1.DeploymentSpec", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_app_v1alpha1_AkkaClusterSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "AkkaClusterSpec defines the desired state of AkkaCluster",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"replicas": {
						SchemaProps: spec.SchemaProps{
							Description: "Number of desired pods. This is a pointer to distinguish between explicit zero and not specified. Defaults to 1.",
							Type:        []string{"integer"},
							Format:      "int32",
						},
					},
					"selector": {
						SchemaProps: spec.SchemaProps{
							Description: "Label selector for pods. Existing ReplicaSets whose pods are selected by this will be the ones affected by this deployment. It must match the pod template's labels.",
							Ref:         ref("k8s.io/apimachinery/pkg/apis/meta/v1.LabelSelector"),
						},
					},
					"template": {
						SchemaProps: spec.SchemaProps{
							Description: "Template describes the pods that will be created.",
							Ref:         ref("k8s.io/api/core/v1.PodTemplateSpec"),
						},
					},
					"strategy": {
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								"x-kubernetes-patch-strategy": "retainKeys",
							},
						},
						SchemaProps: spec.SchemaProps{
							Description: "The deployment strategy to use to replace existing pods with new ones.",
							Ref:         ref("k8s.io/api/apps/v1.DeploymentStrategy"),
						},
					},
					"minReadySeconds": {
						SchemaProps: spec.SchemaProps{
							Description: "Minimum number of seconds for which a newly created pod should be ready without any of its container crashing, for it to be considered available. Defaults to 0 (pod will be considered available as soon as it is ready)",
							Type:        []string{"integer"},
							Format:      "int32",
						},
					},
					"revisionHistoryLimit": {
						SchemaProps: spec.SchemaProps{
							Description: "The number of old ReplicaSets to retain to allow rollback. This is a pointer to distinguish between explicit zero and not specified. Defaults to 10.",
							Type:        []string{"integer"},
							Format:      "int32",
						},
					},
					"paused": {
						SchemaProps: spec.SchemaProps{
							Description: "Indicates that the deployment is paused.",
							Type:        []string{"boolean"},
							Format:      "",
						},
					},
					"progressDeadlineSeconds": {
						SchemaProps: spec.SchemaProps{
							Description: "The maximum time in seconds for a deployment to make progress before it is considered to be failed. The deployment controller will continue to process failed deployments and a condition with a ProgressDeadlineExceeded reason will be surfaced in the deployment status. Note that progress will not be estimated during the time a deployment is paused. Defaults to 600s.",
							Type:        []string{"integer"},
							Format:      "int32",
						},
					},
				},
				Required: []string{"selector", "template"},
			},
		},
		Dependencies: []string{
			"k8s.io/api/apps/v1.DeploymentStrategy", "k8s.io/api/core/v1.PodTemplateSpec", "k8s.io/apimachinery/pkg/apis/meta/v1.LabelSelector"},
	}
}

func schema_pkg_apis_app_v1alpha1_AkkaClusterStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "AkkaClusterStatus defines the observed state of AkkaCluster",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"managementHost": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"managementPort": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
					"lastUpdate": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.Time"),
						},
					},
					"cluster": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/app/v1alpha1.AkkaClusterManagementStatus"),
						},
					},
				},
				Required: []string{"managementHost", "managementPort", "lastUpdate", "cluster"},
			},
		},
		Dependencies: []string{
			"./pkg/apis/app/v1alpha1.AkkaClusterManagementStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.Time"},
	}
}
