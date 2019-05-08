package akkacluster

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	appv1alpha1 "github.com/lightbend/akka-cluster-operator/pkg/apis/app/v1alpha1"
)

var log = logf.Log.WithName("controller_akkacluster")

// Add creates a new AkkaCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAkkaCluster{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("akkacluster-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AkkaCluster
	err = c.Watch(&source.Kind{Type: &appv1alpha1.AkkaCluster{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resources and requeue the owner AkkaCluster
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.AkkaCluster{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileAkkaCluster{}

// ReconcileAkkaCluster reconciles a AkkaCluster object
type ReconcileAkkaCluster struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a AkkaCluster object and makes changes based on the state read
// and what is in the AkkaCluster.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAkkaCluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("name", fmt.Sprintf("%s/%s", request.Namespace, request.Name))
	reqLogger.Info("Reconciling AkkaCluster")

	// Fetch the AkkaCluster instance
	akkaCluster := &appv1alpha1.AkkaCluster{}
	err := r.client.Get(context.TODO(), request.NamespacedName, akkaCluster)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Below are some default installer steps, meaning it'll set things up and that's it.
	// TODO: handle lifecycle changes like updates to spec and scale which need to propagate to deployment.
	// TODO: handle akka cluster status updates.

	// generateResources populates akkaCluster with defaults and returns list of resources to check
	for _, resource := range generateResources(akkaCluster) {
		err = r.client.Get(context.TODO(), request.NamespacedName, resource)
		if err != nil && errors.IsNotFound(err) {
			if err := controllerutil.SetControllerReference(akkaCluster, resource, r.scheme); err != nil {
				return reconcile.Result{}, err
			}
			if err := r.client.Create(context.TODO(), resource); err != nil {
				reqLogger.Info("Tried to create a new resource", "kind", reflect.TypeOf(resource), "error", err)
				return reconcile.Result{}, err
			}
			reqLogger.Info("Creating resource", "kind", reflect.TypeOf(resource))
			return reconcile.Result{Requeue: true}, nil
		}
	}

	// set cluster status
	pods := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(akkaCluster.Spec.Selector.MatchLabels)
	listOps := &client.ListOptions{Namespace: akkaCluster.Namespace, LabelSelector: labelSelector}
	err = r.client.List(context.TODO(), listOps, pods)
	if err != nil {
		reqLogger.Info("error fetching pods", "err", err)
		return reconcile.Result{Requeue: true}, fmt.Errorf("failed to get pods: %v", err)
	}

	findManagementPort := func(pod corev1.Pod) int32 {
		for _, container := range pod.Spec.Containers {
			for _, port := range container.Ports {
				if port.Name == "management" {
					return port.ContainerPort
				}
			}
		}
		return 8558
	}

	currentStatus := appv1alpha1.AkkaClusterStatus{}
	for _, pod := range pods.Items {
		ip := pod.Status.PodIP
		port := findManagementPort(pod)
		if ip == "" || port == 0 {
			continue
		}
		reqLogger.Info("fetching /cluster/members/", "pod", ip, "port", port)
		resp, err := http.Get(fmt.Sprintf("http://%s:%d/cluster/members/", ip, port))
		if err == nil {
			body, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err == nil && json.Unmarshal(body, &currentStatus) == nil {
				// managed to read in someone's cluster status, otherwise keep searching
				reqLogger.Info("found cluster status")
				break
			}
		} else {
			reqLogger.Info("error fetching cluster/members", "err", err)
		}
	}
	if !reflect.DeepEqual(akkaCluster.Status, currentStatus) {
		reqLogger.Info("set cluster status")
		akkaCluster.Status = currentStatus
		err := r.client.Status().Update(context.TODO(), akkaCluster)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}
