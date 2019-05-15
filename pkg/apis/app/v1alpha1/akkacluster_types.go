package v1alpha1

import (
	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

// AkkaClusterMemberStatus corresponds to Akka Management members entries
// ref https://github.com/akka/akka-management/blob/master/cluster-http/src/main/scala/akka/management/cluster/ClusterHttpManagementProtocol.scala
type AkkaClusterMemberStatus struct {
	Node   string   `json:"node"`
	Status string   `json:"status"`
	Roles  []string `json:"roles"`
}

// AkkaClusterUnreachableMemberStatus reports node(s) to node reachability problems
type AkkaClusterUnreachableMemberStatus struct {
	Node       string   `json:"node"`
	ObservedBy []string `json:"observedBy"`
}

// AkkaClusterSpec defines the desired state of AkkaCluster
// +k8s:openapi-gen=true
type AkkaClusterSpec struct {
	apps.DeploymentSpec `json:",inline"`
}

// AkkaClusterStatus defines the observed state of AkkaCluster
// +k8s:openapi-gen=true
type AkkaClusterStatus struct {
	apps.DeploymentStatus `json:",inline"`

	Members       []AkkaClusterMemberStatus            `json:"members"`
	Unreachable   []AkkaClusterUnreachableMemberStatus `json:"unreachable"`
	Leader        string                               `json:"leader"`
	Oldest        string                               `json:"oldest"`
	OldestPerRole map[string]string                    `json:"oldestPerRole"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AkkaCluster is the Schema for the akkaclusters API
// +k8s:openapi-gen=true
type AkkaCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   apps.DeploymentSpec `json:"spec,omitempty"`
	Status AkkaClusterStatus   `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AkkaClusterList contains a list of AkkaCluster
type AkkaClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AkkaCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AkkaCluster{}, &AkkaClusterList{})
}
