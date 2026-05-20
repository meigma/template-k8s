package telemetry

import (
	"fmt"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const EventReasonChildResourcesApplied = "ChildResourcesApplied"

// EventSink is the subset of Kubernetes event recorder behavior this package needs.
type EventSink interface {
	Eventf(object runtime.Object, eventtype string, reason string, messageFmt string, args ...any)
}

// Recorder records controller-specific metrics and Kubernetes Events.
type Recorder interface {
	RecordChildApplies(object runtime.Object, applies []ChildApply)
	RecordStatusTransition(object runtime.Object, condition metav1.Condition)
}

type recorder struct {
	metrics *Metrics
	events  EventSink
}

type noopRecorder struct{}

// NewRecorder combines optional metrics and Kubernetes Events behind one controller dependency.
func NewRecorder(metrics *Metrics, events EventSink) Recorder {
	return recorder{
		metrics: metrics,
		events:  events,
	}
}

// NoopRecorder returns a recorder that intentionally drops all observations.
func NoopRecorder() Recorder {
	return noopRecorder{}
}

func (r recorder) RecordChildApplies(object runtime.Object, applies []ChildApply) {
	summary := newApplySummary()
	for _, apply := range applies {
		if r.metrics != nil {
			if !r.metrics.recordChildApply(apply) {
				continue
			}
		} else if _, ok := metricOperation(apply.Operation); !ok {
			continue
		}
		summary.add(apply)
	}

	if r.events == nil || len(summary.parts) == 0 {
		return
	}
	r.events.Eventf(
		object,
		corev1.EventTypeNormal,
		EventReasonChildResourcesApplied,
		"Applied child resources: %s",
		summary.String(),
	)
}

func (r recorder) RecordStatusTransition(object runtime.Object, condition metav1.Condition) {
	if r.metrics != nil {
		r.metrics.recordStatusTransition(condition)
	}
	if r.events == nil {
		return
	}
	r.events.Eventf(
		object,
		corev1.EventTypeNormal,
		condition.Reason,
		"%s condition is %s: %s",
		condition.Type,
		condition.Status,
		condition.Message,
	)
}

func (noopRecorder) RecordChildApplies(runtime.Object, []ChildApply) {}

func (noopRecorder) RecordStatusTransition(runtime.Object, metav1.Condition) {}

type applySummary struct {
	parts []string
}

func newApplySummary() applySummary {
	return applySummary{
		parts: make([]string, 0, 3),
	}
}

func (s *applySummary) add(apply ChildApply) {
	operation, _ := metricOperation(apply.Operation)
	s.parts = append(s.parts, fmt.Sprintf("%s %s", childResourceEventName(apply.Resource), operation))
}

func (s applySummary) String() string {
	sort.Strings(s.parts)
	return strings.Join(s.parts, ", ")
}

func childResourceEventName(resource ChildResource) string {
	switch resource {
	case ChildResourceConfigMap:
		return "ConfigMap"
	case ChildResourceDeployment:
		return "Deployment"
	case ChildResourceService:
		return "Service"
	default:
		return string(resource)
	}
}
