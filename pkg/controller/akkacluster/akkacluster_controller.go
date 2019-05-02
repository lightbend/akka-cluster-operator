package akkacluster

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	appv1alpha1 "github.com/lightbend/akka-cluster-operator/pkg/apis/app/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
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

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner AkkaCluster
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
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

	// helper function to call create and requeue for next step
	createResource := func(obj runtime.Object) (reconcile.Result, error) {
		if err := r.client.Create(context.TODO(), obj); err != nil {
			reqLogger.Info("Tried to create a new resource", "kind", reflect.TypeOf(obj), "error", err)
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating resource", "kind", reflect.TypeOf(obj))
		return reconcile.Result{Requeue: true}, nil
	}

	// create service account
	serviceAccount := &corev1.ServiceAccount{}
	err = r.client.Get(context.TODO(), request.NamespacedName, serviceAccount)
	if err != nil && errors.IsNotFound(err) {
		serviceAccount.Name = akkaCluster.Name
		serviceAccount.Namespace = akkaCluster.Namespace
		if err := controllerutil.SetControllerReference(akkaCluster, serviceAccount, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		return createResource(serviceAccount)
	}

	// create pod-reader role
	role := &rbac.Role{}
	err = r.client.Get(context.TODO(), request.NamespacedName, role)
	if err != nil && errors.IsNotFound(err) {
		role.Name = akkaCluster.Name
		role.Namespace = akkaCluster.Namespace
		role.Rules = []rbac.PolicyRule{
			rbac.PolicyRule{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "watch", "list"},
			},
		}
		if err := controllerutil.SetControllerReference(akkaCluster, role, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		return createResource(role)
	}

	// create role binding
	roleBinding := &rbac.RoleBinding{}
	err = r.client.Get(context.TODO(), request.NamespacedName, roleBinding)
	if err != nil && errors.IsNotFound(err) {
		roleBinding.Name = akkaCluster.Name
		roleBinding.Namespace = akkaCluster.Namespace
		roleBinding.RoleRef = rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     role.Kind,
			Name:     role.Name,
		}
		roleBinding.Subjects = []rbac.Subject{
			rbac.Subject{
				Kind: serviceAccount.Kind,
				Name: serviceAccount.Name,
			},
		}
		if err := controllerutil.SetControllerReference(akkaCluster, roleBinding, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		return createResource(roleBinding)
	}

	// create deployment
	deployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), request.NamespacedName, deployment)
	if err != nil && errors.IsNotFound(err) {
		deployment.Name = akkaCluster.Name
		deployment.Namespace = akkaCluster.Namespace
		deployment.Spec = akkaCluster.Spec

		deployment.Spec.Template.Spec.ServiceAccountName = serviceAccount.Name

		makeSet := func(m map[string]string, k string, v string) map[string]string {
			if m == nil {
				m = make(map[string]string)
			}
			m[k] = v
			return m
		}
		deployment.Labels = makeSet(deployment.Labels, "app", akkaCluster.Name) // needed?
		if deployment.Spec.Selector == nil {
			deployment.Spec.Selector = &metav1.LabelSelector{}
		}
		deployment.Spec.Selector.MatchLabels = makeSet(deployment.Spec.Selector.MatchLabels, "app", akkaCluster.Name)
		deployment.Spec.Template.Labels = makeSet(deployment.Spec.Template.Labels, "app", akkaCluster.Name)

		deployment.Spec.Strategy.Type = appsv1.RollingUpdateDeploymentStrategyType
		one := intstr.FromInt(1)
		zero := intstr.FromInt(0)
		if deployment.Spec.Strategy.RollingUpdate == nil {
			deployment.Spec.Strategy.RollingUpdate = &appsv1.RollingUpdateDeployment{}
		}
		deployment.Spec.Strategy.RollingUpdate.MaxSurge = &one
		deployment.Spec.Strategy.RollingUpdate.MaxUnavailable = &zero

		for i := range deployment.Spec.Template.Spec.Containers {
			deployment.Spec.Template.Spec.Containers[i].Env = append(deployment.Spec.Template.Spec.Containers[i].Env,
				corev1.EnvVar{
					Name:  "AKKA_CLUSTER_BOOTSTRAP_SERVICE_NAME",
					Value: akkaCluster.Name,
				},
				// TODO CONTACT_PT_NR
			)
		}

		if err := controllerutil.SetControllerReference(akkaCluster, deployment, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		return createResource(deployment)
	}

	// set cluster status
	pods := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(deployment.Spec.Selector.MatchLabels)
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
