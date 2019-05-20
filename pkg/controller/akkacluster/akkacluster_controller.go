package akkacluster

import (
	"context"
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	appv1alpha1 "github.com/lightbend/akka-cluster-operator/pkg/apis/app/v1alpha1"
)

var log = logf.Log.WithName("controller_akkacluster")

// Add creates a new AkkaCluster Controller and adds it to the Manager. The Manager will
// set fields on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	statusEvents := make(chan event.GenericEvent, 1024)
	apiClient := mgr.GetClient()

	r := &ReconcileAkkaCluster{
		client:      apiClient,
		scheme:      mgr.GetScheme(),
		events:      statusEvents,
		statusActor: NewStatusActor(statusEvents, apiClient),
	}

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

	// Watch for changes to all possible secondary resources
	for _, obj := range allPossibleGeneratedResourceTypes() {
		err = c.Watch(&source.Kind{Type: obj}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &appv1alpha1.AkkaCluster{},
		})
		if err != nil {
			return err
		}
	}

	// watch for Akka Cluster status updates
	err = c.Watch(&source.Channel{Source: r.events}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// watch for pod events to trigger status updates
	// TODO: is this needed or are these percolating up to Deployment well enough?
	// need to test with and without readiness, as they may be equivalent only with readiness
	c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: false,
		OwnerType:    &appv1alpha1.AkkaCluster{},
	})
	if err != nil {
		return err
	}

	// debug watchers (this could be redone as debug predicates)
	// c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &enqueueDebugger{})
	// c.Watch(&source.Kind{Type: &corev1.Pod{}}, &enqueueDebugger{})
	// c.Watch(&source.Kind{Type: &appv1alpha1.AkkaCluster{}}, &enqueueDebugger{})

	return nil
}

var _ reconcile.Reconciler = &ReconcileAkkaCluster{}

// ReconcileAkkaCluster reconciles a AkkaCluster object
type ReconcileAkkaCluster struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client      client.Client
	scheme      *runtime.Scheme
	events      chan event.GenericEvent
	statusActor *StatusActor
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
			r.statusActor.StopPolling(request)
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// generateResources populates akkaCluster with defaults and returns list of resources to check.
	for _, wantedResource := range generateResources(akkaCluster) {
		if err := controllerutil.SetControllerReference(akkaCluster, wantedResource, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		kind := reflect.ValueOf(wantedResource).Elem().Type().String()
		// Fetch this resource from cluster, if any.
		clusterResource := wantedResource.DeepCopyObject()
		err = r.client.Get(context.TODO(), request.NamespacedName, clusterResource)
		if err != nil && errors.IsNotFound(err) {
			// Create wanted resource. Next client.Get will at least fetch wantedResource and will eventually
			// reflect the object as it is in the cluster.
			if err := r.client.Create(context.TODO(), wantedResource); err != nil {
				reqLogger.Info("Tried to create a new resource", "kind", kind, "error", err)
				return reconcile.Result{}, err
			}
			reqLogger.Info("Creating resource", "kind", kind)
			return reconcile.Result{Requeue: true}, nil
		}
		// Update cluster resource to wanted resource, if needed.
		if !SubsetEqual(wantedResource, clusterResource) {
			reqLogger.Info("applying update", "kind", kind, "match")

			if err := r.client.Update(context.TODO(), wantedResource); err != nil {
				reqLogger.Info("Tried to update resource", "kind", kind, "error", err)
				return reconcile.Result{}, err
			}
			return reconcile.Result{Requeue: true}, nil
		}
	}

	if r.statusActor != nil {
		currentStatus := r.statusActor.GetStatus(request)
		if nil != currentStatus {
			if !reflect.DeepEqual(akkaCluster.Status, currentStatus) {
				akkaCluster.Status = currentStatus
				err := r.client.Status().Update(context.TODO(), akkaCluster)
				if err != nil {
					reqLogger.Info("update error", "err", err)
					return reconcile.Result{}, err
				}
				reqLogger.Info("updated cluster status")
			}
		}
		// StartPolling means: notify me if status for this cluster changes from what I've
		// got so far. This could happen on the first reconcile, meaning status is unknown
		// or nil, and we want to be notified when it becomes available. This could also
		// happen on a reconcile triggered by a status update, in which case we want to
		// first GetStatus() and update the cluster object, then poll for future change.
		r.statusActor.StartPolling(akkaCluster)
	}

	return reconcile.Result{}, nil
}
