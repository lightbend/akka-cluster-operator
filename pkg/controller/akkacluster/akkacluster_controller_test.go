package akkacluster

import (
	"context"
	"encoding/json"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	appv1alpha1 "github.com/lightbend/akka-cluster-operator/pkg/apis/app/v1alpha1"
)

func TestAkkaController(t *testing.T) {
	name := types.NamespacedName{
		Name:      "akka-cluster-test",
		Namespace: "akka-cluster-namespace",
	}
	clusterJSON := `{
		"apiVersion": "app.lightbend.com/v1alpha1",
		"kind": "AkkaCluster",
		"spec": {
		  "replicas": 3,
		  "template": {
			"spec": {
			  "containers": [
				{
				  "name": "main",
				  "image": "akka-cluster:1.0.0"
				}
			  ]
			}
		  }
		}
	  }`

	akkaCluster := &appv1alpha1.AkkaCluster{}
	json.Unmarshal([]byte(clusterJSON), akkaCluster)
	akkaCluster.ObjectMeta.Name = name.Name
	akkaCluster.ObjectMeta.Namespace = name.Namespace

	// mock context for reconciler
	scheme := scheme.Scheme
	scheme.AddKnownTypes(appv1alpha1.SchemeGroupVersion, akkaCluster)
	client := fake.NewFakeClientWithScheme(scheme, akkaCluster)
	r := &ReconcileAkkaCluster{client: client, scheme: scheme}

	// mock event loop
	req := reconcile.Request{NamespacedName: name}
	eventLoop := func() {
		limit := 10
		for ; limit > 0; limit-- {
			res, err := r.Reconcile(req)
			if err != nil {
				t.Fatalf("reconcile error: %v", err)
			}
			if res.Requeue != true || res.RequeueAfter != 0 {
				break
			}
		}
		if limit <= 0 {
			t.Fatalf("reconcile didn't resolve within expected number of passes")
		}
	}
	eventLoop()

	// check that expected resources were created
	serviceAccount := &corev1.ServiceAccount{}
	err := client.Get(context.TODO(), req.NamespacedName, serviceAccount)
	if err != nil {
		t.Error(err)
	}

	role := &rbac.Role{}
	err = client.Get(context.TODO(), req.NamespacedName, role)
	if err != nil {
		t.Error(err)
	}

	rolebinding := &rbac.RoleBinding{}
	err = client.Get(context.TODO(), req.NamespacedName, rolebinding)
	if err != nil {
		t.Error(err)
	}

	deployment := &appsv1.Deployment{}
	err = client.Get(context.TODO(), req.NamespacedName, deployment)
	if err != nil {
		t.Error(err)
	}
	if *deployment.Spec.Replicas != 3 {
		t.Errorf("expected three replicas but got %d", deployment.Spec.Replicas)
	}

	// grow the cluster
	*akkaCluster.Spec.Replicas = 4
	client.Update(context.TODO(), akkaCluster)
	eventLoop()

	err = client.Get(context.TODO(), req.NamespacedName, deployment)
	if err != nil {
		t.Error(err)
	}
	if *deployment.Spec.Replicas != 4 {
		t.Errorf("expected four replicas but got %d", deployment.Spec.Replicas)
	}

}
