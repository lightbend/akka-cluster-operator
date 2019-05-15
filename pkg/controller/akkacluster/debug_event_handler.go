package akkacluster

import (
	"encoding/json"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type enqueueDebugger struct {
}

// TODO: might be more clever to use a debugger predicate on the real watch instead of these fake watches
var elog = logf.Log.WithName("EventDebugger")

func (*enqueueDebugger) Create(e event.CreateEvent, _ workqueue.RateLimitingInterface) {
	elog.Info("Create", "uri", e.Meta.GetSelfLink())
}

// Update is called in response to an update event -  e.g. Pod Updated.
func (*enqueueDebugger) Update(e event.UpdateEvent, _ workqueue.RateLimitingInterface) {
	elog.Info("Update", "uir", e.MetaNew.GetSelfLink())
	// b, _ := json.Marshal(&e)
	// fmt.Println("Update", string(b))
}

// Delete is called in response to a delete event - e.g. Pod Deleted.
func (*enqueueDebugger) Delete(e event.DeleteEvent, _ workqueue.RateLimitingInterface) {
	elog.Info("Delete", "uri", e.Meta.GetSelfLink(), "deleteStateUnknown", e.DeleteStateUnknown)
	// b, _ := json.Marshal(&e)
	// fmt.Println("Delete", string(b))
}

// Generic is called in response to an event of an unknown type or a synthetic event triggered as a cron or
// external trigger request - e.g. reconcile Autoscaling, or a Webhook.
func (*enqueueDebugger) Generic(e event.GenericEvent, _ workqueue.RateLimitingInterface) {
	b, _ := json.Marshal(&e)
	elog.Info("Generic", "uri", e.Meta.GetSelfLink(), "json", string(b))
}
