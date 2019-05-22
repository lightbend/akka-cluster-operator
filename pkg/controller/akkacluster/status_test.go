package akkacluster

import (
	"encoding/json"
	"math/rand"
	"net/url"
	"strconv"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appv1alpha1 "github.com/lightbend/akka-cluster-operator/pkg/apis/app/v1alpha1"
)

// In testing a StatusActor we want to set context of a podLister and urlReader suitable
// for simulating various scenarios.

// akkacluster_types has the client schema, here we extend that to server schema :-|
// essentially extra fields the operator client is supposed to ignore.

type testMemberStatus struct {
	appv1alpha1.AkkaClusterMemberStatus `json:",inline"`
	NodeUID                             string `json:"nodeUid"`
}

type testManagementStatus struct {
	Members       []testMemberStatus                               `json:"members"`
	Unreachable   []appv1alpha1.AkkaClusterUnreachableMemberStatus `json:"unreachable"`
	Leader        string                                           `json:"leader"`
	Oldest        string                                           `json:"oldest"`
	OldestPerRole map[string]string                                `json:"oldestPerRole"`
	SelfNode      string                                           `json:"selfNode"`
}

// generateNodeStatus returns a random Akka node status value. Values can be excluded, so
// one may call generateNodeStatusExcept("Up", "WeaklyUp") to exclude those two values.
func generateNodeStatus(not ...string) string {
	statuses := []string{"Joining", "Up", "Leaving", "Exiting", "Down", "Removed", "WeaklyUp"}
	for _, remove := range not {
		for n := range statuses {
			if remove == statuses[n] {
				statuses = append(statuses[:n], statuses[n+1:]...)
				break
			}
		}
	}
	return statuses[rand.Intn(len(statuses))]
}
func generateNodestatusExcept(not ...string) string {
	return generateNodeStatus(not...)
}
func generateNodestatusExceptUp() string {
	return generateNodeStatus("Up")
}

// generateNodeUID is a random long.
func generateNodeUID() string {
	return strconv.FormatInt(int64(rand.Uint64()), 10)
}

// node address is "akka.tcp://actorSystem@host:port" where ".tcp" is optional and "host"
// varies between nodes but the rest stays the same. protocol://system@host:port
func generateNodeAddress(host string) string {
	return "akka.tcp://someActorSystem@" + host + ":2552"
}

// mock pods need specified values in
// pod.Status.PodIP
// pod.DeletionTimestamp
// pod.Status.Phase    // Pending, Running, Succeeded, Failed, Unknown
// pod.Spec.Containers[].Ports[]
func generatePod(ip string) *corev1.Pod {
	pod := &corev1.Pod{}
	pod.Status.PodIP = ip
	pod.Status.Phase = corev1.PodRunning
	pod.Spec.Containers = []corev1.Container{
		{Ports: []corev1.ContainerPort{
			{Name: "management", ContainerPort: 8558},
		}},
	}
	return pod
}

func generateMember(ip string) *testMemberStatus {
	status := &testMemberStatus{
		AkkaClusterMemberStatus: appv1alpha1.AkkaClusterMemberStatus{
			Node:   generateNodeAddress(ip),
			Status: "Up",
			Roles:  []string{"dc"},
		},
		NodeUID: generateNodeUID(),
	}
	return status
}

func generateManagementResult(ips []string) *testManagementStatus {
	status := &testManagementStatus{}
	leader := len(ips) - 1 // rand.Intn(len(ips))
	oldest := rand.Intn(len(ips))
	status.Leader = generateNodeAddress(ips[leader])
	status.Oldest = generateNodeAddress(ips[oldest])
	status.OldestPerRole = map[string]string{"dc": status.Oldest}

	status.Members = []testMemberStatus{}
	for n := range ips {
		status.Members = append(status.Members, *generateMember(ips[n]))
	}

	return status
}

// testReader is a urlReader and podLister mock
type testReaderLister struct {
	ips []string
}

func (r *testReaderLister) ReadURL(uri string) ([]byte, error) {
	status := generateManagementResult(r.ips)
	link, _ := url.Parse(uri)
	status.SelfNode = generateNodeAddress(link.Hostname())
	return json.Marshal(status)
}

func (r *testReaderLister) ListPods(cluster *appv1alpha1.AkkaCluster) *corev1.PodList {
	list := &corev1.PodList{}
	list.Items = []corev1.Pod{}
	for n := range r.ips {
		list.Items = append(list.Items, *generatePod(r.ips[n]))
	}
	return list
}

func TestStatusActor(t *testing.T) {
	statusChanged := make(chan event.GenericEvent)
	ips := []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}
	mock := &testReaderLister{ips}

	actor := &StatusActor{
		inbox:         make(chan func(), 100),
		statusChanged: statusChanged,
		lister:        mock,
		reader:        mock,
		minimalWait:   time.Millisecond,
		polls:         make(map[reconcile.Request]pollingRequest),
	}
	go actor.Run()

	cluster := &appv1alpha1.AkkaCluster{}
	cluster.Name = "boop"
	cluster.Namespace = "bop"
	actor.StartPolling(cluster)
	<-statusChanged
	status := actor.GetStatus(getReq(cluster))
	if len(status.Cluster.Members) != len(mock.ips) {
		t.Errorf("expected %d cluster members but got %d", len(mock.ips), len(status.Cluster.Members))
	}

	cluster.Status = status
	actor.StartPolling(cluster)
	<-statusChanged
	status = actor.GetStatus(getReq(cluster))
	if status.ManagementHost != "10.0.0.3" {
		t.Errorf("expected management host to converge on leader, but got %s", status.ManagementHost)
	}

}
