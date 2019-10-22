package akkacluster

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appv1alpha1 "github.com/lightbend/akka-cluster-operator/pkg/apis/app/v1alpha1"
)

//
// On status liveness:
//
// The AkkaCluster operator tries to report the Akka Management status from the point of
// view of the leader of the cluster, and keep it reasonably lively. To do this, it checks
// status on the leader whenever there is a reconciliation event. The challenge then is
// Kubernetes resource mutations might precede Akka Cluster events. For example a new pod
// triggers a controller reconciliation, but only later after the app has started will it
// join the Akka cluster. So we might typically not see an Akka status change right away
// after pods have changed, but if we checked a few seconds later we'd see a change. To
// handle this, we poll Akka status for a bit after reconciliation. If we see status
// change, we signal for reconciliation and so in turn start polling again.
//
// To do this, we have an actor dedicated to looking at Akka cluster status. It takes a
// signal to start polling a given cluster with a known status, and in turn it signals
// back if status changes on any of the polls. The controller-runtime mechanism for the
// signal back is via source.Channel and GenericEvent. As an optimization, this actor is
// also a provider of cluster status, so Reconcile() just reads the result instead of
// rebuilding it from scratch. Mechanically these two functions, get() and poll(), work
// together as a step. The Controller calls get() to step up to the latest status, then
// poll() to slide over to the next step, climbing status changes over time.
//
// In future it might be possible to leverage Server Sent Events from the Akka Cluster as
// a replacement for polling. https://github.com/akka/akka-management/issues/540
//

// Given a URL, return the body of the response.
type urlReader interface {
	ReadURL(string) ([]byte, error)
}

// An httpReader is a urlReader with http.Client.
type httpReader struct {
	http.Client
}

func newHTTPReader() *httpReader {
	return &httpReader{
		http.Client{Timeout: 3 * time.Second},
	}
}

func (r *httpReader) ReadURL(url string) ([]byte, error) {
	resp, err := r.Get(url)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return body, err
}

// Given an AkkaCluster, return a list of pods.
type podLister interface {
	ListPods(*appv1alpha1.AkkaCluster) *corev1.PodList
}

// controllerPodLister is a podLister with a controller client
type controllerPodLister struct {
	client.Client
}

func (p *controllerPodLister) ListPods(cluster *appv1alpha1.AkkaCluster) *corev1.PodList {
	pods := &corev1.PodList{}
	listOps := &client.ListOptions{
		Namespace:     cluster.Namespace,
		LabelSelector: labels.SelectorFromSet(cluster.Spec.Selector.MatchLabels),
	}
	p.List(context.TODO(), listOps, pods)
	return pods
}

// StatusActor manages updating status for a set of Akka clusters. It is a worker for a
// controller, responsible mainly for converting Akka cluster events into controller
// reconciliation events. It also provides cluster status to the controller.
type StatusActor struct {
	// inbound:
	inbox chan func()
	// outbound:
	statusChanged chan event.GenericEvent
	lister        podLister
	reader        urlReader
	// state:
	minimalWait time.Duration
	polls       map[reconcile.Request]pollingRequest
}

type pollingRequest struct {
	cluster    *appv1alpha1.AkkaCluster
	waitFactor int
	timer      *time.Timer
}

// NewStatusActor constructs a new StatusActor given a Manager's api client and some
// channel for status update events.
func NewStatusActor(client client.Client, statusChanged chan event.GenericEvent) *StatusActor {
	actor := &StatusActor{
		inbox:         make(chan func(), 100),
		statusChanged: statusChanged,
		lister:        &controllerPodLister{client},
		reader:        newHTTPReader(),
		minimalWait:   time.Second,
		polls:         make(map[reconcile.Request]pollingRequest),
	}
	go actor.Run()
	return actor
}

// Run loops on inbox until it is closed.
func (a *StatusActor) Run() {
	for f := range a.inbox {
		f()
	}
}

// getReq assembles a NamespacedName from metadata of AkkaCluster.
func getReq(cluster *appv1alpha1.AkkaCluster) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
	}
}

// StartPolling tries to update status of a given cluster for some limited amount of time.
// If called before existing polling is finished, it will stop previous and restart. A
// copy of the AkkaCluster object is made in the caller's frame, which is then inserted
// into the actor's work state in the actor's frame.
//
// This work request is quiescing, which is a kind of work rate limiter. Practically this
// means many outside callers can ask to start polling an arbitrary number of times, but
// they'll cancel each other if within the minimal wait window (eg 1 second). This means
// the first poll for status will only happen if there is a quiet window with only one
// request in the last second. This fits the typical reconcile burst, which may send
// dozens of update requests per second, and here we wait until the burst seems to be
// over. This would not work well if requests for polling were continuous to the point
// where no quiet period happened.
func (a *StatusActor) StartPolling(cluster *appv1alpha1.AkkaCluster) {
	reqKey := getReq(cluster)
	copy := cluster.DeepCopy()
	isImmediate := make(chan bool)

	a.inbox <- func() {
		poll, ok := a.polls[reqKey]
		immediateMode := true
		if ok && poll.timer != nil {
			poll.timer.Stop()
			immediateMode = false // quiesce requests for minimalWait
		}
		msg := pollingRequest{
			cluster: copy,
		}
		if !immediateMode {
			// restart previous poll after minimalWait
			msg.timer = time.AfterFunc(a.minimalWait, func() { a.update(reqKey) })
		}
		a.polls[reqKey] = msg
		isImmediate <- immediateMode
	}

	if <-isImmediate {
		a.update(reqKey)
	}
}

// GetStatus looks up last known status by name and namespace. A nil result means no known
// last status, so caller may safely ignore nil results and hope for better next time.
func (a *StatusActor) GetStatus(req reconcile.Request) *appv1alpha1.AkkaClusterStatus {
	status := make(chan *appv1alpha1.AkkaClusterStatus)
	a.inbox <- func() {
		poll, ok := a.polls[req]
		if !ok {
			log.Info("StatusActor asked for missing status", "name", req.String())
			status <- nil
		} else {
			status <- poll.cluster.Status.DeepCopy()
		}
	}
	return <-status
}

// StopPolling stops timer and removes polling state for a given cluster. This is optional
// since polling against a removed cluster will stop trying and remove itself eventually.
func (a *StatusActor) StopPolling(req reconcile.Request) {
	a.inbox <- func() {
		poll, ok := a.polls[req]
		if ok {
			if poll.timer != nil {
				poll.timer.Stop()
			}
			delete(a.polls, req)
		}
	}
}

// update process
// 1. try Leader, otherwise get random Pod IP and try that
// 2. if status is different, save status and signal statusChanged
// 3. otherwise double the wait time and retry up to some limit
func (a *StatusActor) update(req reconcile.Request) {
	a.inbox <- func() {
		poll, ok := a.polls[req]
		if !ok {
			return
		}
		a.initStatus(poll.cluster)
		currentStatus := a.fetchUpdate(poll.cluster)

		if currentStatus == nil {
			// write something so that initial status is not nil
			poll.cluster.Status.LastUpdate = metav1.Now()
			// start from scratch next time, maybe picking different pod
			poll.cluster.Status.ManagementHost = ""
		} else if !reflect.DeepEqual(currentStatus.Cluster, poll.cluster.Status.Cluster) {
			// found a change: save it, signal upstream, stop polling
			poll.cluster.Status = currentStatus
			poll.cluster.Status.LastUpdate = metav1.Now()
			poll.timer = nil
			a.polls[req] = poll
			a.statusChanged <- event.GenericEvent{
				Meta:   poll.cluster,
				Object: poll.cluster,
			}
			return
		}
		// poll again, up to some limit
		if poll.waitFactor == 0 {
			poll.waitFactor = 1
		}
		poll.waitFactor *= 2
		if poll.waitFactor > 60 {
			// State is stored in the AkkaCluster object, so we can be parsimonious here.
			// A new update request will pick up where it left off with previous host and
			// port etc since those are in the request object at this point.
			delete(a.polls, req)
			return
		}
		poll.timer = time.AfterFunc(a.minimalWait*time.Duration(poll.waitFactor), func() { a.update(req) })
		a.polls[req] = poll
	}
}

// initStatus sets ManagementHost and ManagementPort as needed. We re-use the Leader or
// ManagementHost from before. Otherwise it will use the API client to find a running pod
// and use that as a starting point. Clearing ManagementHost before calling initStatus
// will cause it to find a running pod, as a way of starting over when last try failed.
func (a *StatusActor) initStatus(cluster *appv1alpha1.AkkaCluster) error {
	if cluster.Status == nil {
		cluster.Status = &appv1alpha1.AkkaClusterStatus{}
	}
	// Coalesce on the Leader for updates, if things are working. The strong assumption
	// made here is that if leaderURL is unworkable, then the cluster isn't workable. So
	// here we presume the difference between a PodIP and leader.Hostname() is not
	// something we need to probe and confirm, and can use either interchangably. And in
	// practice it looks like typical hostnames here are just the same as pod ips.
	if cluster.Status.ManagementHost != "" {
		leaderURL, _ := url.Parse(cluster.Status.Cluster.Leader)
		if leaderURL.Hostname() != "" {
			cluster.Status.ManagementHost = leaderURL.Hostname()
		}
	}
	if cluster.Status.ManagementHost == "" {
		pod := a.findRunningPod(cluster)
		if nil == pod {
			return errors.New("no running cluster members")
		}
		cluster.Status.ManagementHost = pod.Status.PodIP
		if cluster.Status.ManagementPort == 0 {
			cluster.Status.ManagementPort = findManagementPort(pod)
		}
	}
	return nil
}

// fetchUpdate does a bunch of IO that can fail on bad config, network failures, http
// failure, parsing failure. It returns a status object if found. No distinction is made
// amongst the various errors, which are all presumed to perhaps work in the future.
func (a *StatusActor) fetchUpdate(cluster *appv1alpha1.AkkaCluster) *appv1alpha1.AkkaClusterStatus {
	if cluster.Status.ManagementHost == "" {
		return nil
	}
	link := fmt.Sprintf("http://%s:%d/cluster/members/",
		cluster.Status.ManagementHost,
		cluster.Status.ManagementPort)
	log.Info("fetching status", "name", cluster.Namespace+"/"+cluster.Name, "url", link)
	body, err := a.reader.ReadURL(link)
	if err != nil {
		log.Info("StatusActor could not read endpoint", "err", err)
		return nil
	}
	currentStatus := cluster.Status.DeepCopy()
	err = json.Unmarshal(body, &currentStatus.Cluster)
	if err != nil {
		return nil
	}
	return currentStatus
}

func findManagementPort(pod *corev1.Pod) int32 {
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			if port.Name == "management" {
				return port.ContainerPort
			}
		}
	}
	return 8558
}

// findRunningPod emulates Akka Management by using the spec Selector to list Pods, then
// filtering on those that have an IP, are not marked for deletion, and currently running.
// This function also shuffles the list of pods to better avoid getting stuck in a loop
// against a running pod without a working management endpoint.
func (a *StatusActor) findRunningPod(cluster *appv1alpha1.AkkaCluster) *corev1.Pod {
	log.Info("fetching pods", "name", cluster.Namespace+"/"+cluster.Name)
	pods := a.lister.ListPods(cluster)
	for n := range rand.Perm(len(pods.Items)) {
		pod := &pods.Items[n]
		if pod.Status.PodIP != "" && pod.DeletionTimestamp == nil && pod.Status.Phase == corev1.PodRunning {
			return pod
		}
	}
	log.Info("no pods found", "name", cluster.Namespace+"/"+cluster.Name)
	return nil
}
