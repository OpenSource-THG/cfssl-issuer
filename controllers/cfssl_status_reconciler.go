package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	cfsslv1beta1 "github.com/opensource-thg/cfssl-issuer/api/v1beta1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type cfsslStatusReconciler struct {
	*CfsslIssuerReconciler
	issuer *cfsslv1beta1.CfsslIssuer
	logger logr.Logger
}

func newCfsslStatusReconciler(r *CfsslIssuerReconciler, iss *cfsslv1beta1.CfsslIssuer, log logr.Logger) *cfsslStatusReconciler {
	return &cfsslStatusReconciler{
		CfsslIssuerReconciler: r,
		issuer:                iss,
		logger:                log,
	}
}

func (r *cfsslStatusReconciler) Update(ctx context.Context, status cfsslv1beta1.ConditionStatus, reason, message string, args ...interface{}) error {
	completeMessage := fmt.Sprintf(message, args...)
	r.setCondition(status, reason, completeMessage)

	// Fire an Event to additionally inform users of the change
	eventType := core.EventTypeNormal
	if status == cfsslv1beta1.ConditionFalse {
		eventType = core.EventTypeWarning
	}
	r.Recorder.Event(r.issuer, eventType, reason, completeMessage)

	if err := r.Client.Update(ctx, r.issuer); err != nil {
		return err
	}

	if err := r.Client.Status().Update(ctx, r.issuer); err != nil {
		return err
	}

	return nil
}

// setCondition will set a 'condition' on the given cfsslv1beta1.CfsslIssuer resource.
//
// - If no condition of the same type already exists, the condition will be
//   inserted with the LastTransitionTime set to the current time.
// - If a condition of the same type and state already exists, the condition
//   will be updated but the LastTransitionTime will not be modified.
// - If a condition of the same type and different state already exists, the
//   condition will be updated and the LastTransitionTime set to the current
//   time.
func (r *cfsslStatusReconciler) setCondition(status cfsslv1beta1.ConditionStatus, reason, message string) {
	now := meta.NewTime(r.Clock.Now())
	c := cfsslv1beta1.CfsslIssuerCondition{
		Type:               cfsslv1beta1.ConditionReady,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: &now,
	}

	// Search through existing conditions
	for idx, cond := range r.issuer.Status.Conditions {
		// Skip unrelated conditions
		if cond.Type != cfsslv1beta1.ConditionReady {
			continue
		}

		// If this update doesn't contain a state transition, we don't update
		// the conditions LastTransitionTime to Now()
		if cond.Status == status {
			c.LastTransitionTime = cond.LastTransitionTime
		} else {
			r.logger.Info("found status change for CfsslIssuer condition; setting lastTransitionTime", "condition", cond.Type, "old_status", cond.Status, "new_status", status, "time", now.Time)
		}

		// Overwrite the existing condition
		r.issuer.Status.Conditions[idx] = c
		return
	}

	// If we've not found an existing condition of this type, we simply insert
	// the new condition into the slice.
	r.issuer.Status.Conditions = append(r.issuer.Status.Conditions, c)
	r.logger.Info("setting lastTransitionTime for CfsslIssuer condition", "condition", cfsslv1beta1.ConditionReady, "time", now.Time)
}
