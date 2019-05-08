package akkacluster

import (
	"encoding/json"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	appv1alpha1 "github.com/lightbend/akka-cluster-operator/pkg/apis/app/v1alpha1"
)

// GenericResource have both meta and runtime interfaces
type GenericResource interface {
	metav1.Object
	runtime.Object
}

// generateResources() produces a list of rbac and deployment resources suitable for akkaCluster.
// If akkaCluster resource does not specify needed options, we provide defaults.
func generateResources(akkaCluster *appv1alpha1.AkkaCluster) []GenericResource {
	resources := []GenericResource{}

	// if akkaCluster has no serviceAccount, generate rbac resources
	if akkaCluster.Spec.Template.Spec.ServiceAccountName == "" {
		// serviceAccount
		serviceAccount := &corev1.ServiceAccount{}
		serviceAccount.Name = akkaCluster.Name
		serviceAccount.Namespace = akkaCluster.Namespace

		// role
		role := &rbac.Role{}
		role.Name = akkaCluster.Name
		role.Namespace = akkaCluster.Namespace
		role.Rules = []rbac.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "watch", "list"},
			},
		}

		// rolebinding
		roleBinding := &rbac.RoleBinding{}
		roleBinding.Name = akkaCluster.Name
		roleBinding.Namespace = akkaCluster.Namespace
		roleBinding.RoleRef = rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     role.Kind,
			Name:     role.Name,
		}
		roleBinding.Subjects = []rbac.Subject{
			{
				Kind: serviceAccount.Kind,
				Name: serviceAccount.Name,
			},
		}

		// connect to pod spec
		akkaCluster.Spec.Template.Spec.ServiceAccountName = serviceAccount.Name

		// enqueue rbac resources for creation later
		resources = append(resources, serviceAccount, role, roleBinding)
	}

	// default label selector, if none given
	if akkaCluster.Spec.Selector == nil {
		selectorKey := "app"
		akkaCluster.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				selectorKey: akkaCluster.Name,
			},
		}
		if akkaCluster.Spec.Template.Labels == nil {
			akkaCluster.Spec.Template.Labels = make(map[string]string)
		}
		akkaCluster.Spec.Template.Labels[selectorKey] = akkaCluster.Name
	}

	// default strategy, if none given
	if akkaCluster.Spec.Strategy.Type == "" {
		// use json serializer here as easier to read
		json.Unmarshal([]byte(`{
			"type": "`+appsv1.RollingUpdateDeploymentStrategyType+`",
			"rollingUpdate": {
				"maxSurge": 1,
				"maxUnavailable": 0
			}
		}`), &akkaCluster.Spec.Strategy)
	}

	// env settings
	for i := range akkaCluster.Spec.Template.Spec.Containers {
		akkaCluster.Spec.Template.Spec.Containers[i].Env = append(akkaCluster.Spec.Template.Spec.Containers[i].Env,
			corev1.EnvVar{
				Name:  "AKKA_CLUSTER_BOOTSTRAP_SERVICE_NAME",
				Value: akkaCluster.Name,
			},
			// TODO CONTACT_PT_NR
		)
	}

	// set up deployment spec
	deployment := &appsv1.Deployment{}
	deployment.Name = akkaCluster.Name
	deployment.Namespace = akkaCluster.Namespace
	deployment.Spec = akkaCluster.Spec

	resources = append(resources, deployment)

	return resources
}
