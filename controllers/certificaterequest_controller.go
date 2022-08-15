/*
Copyright 2019 The cert-manager authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	cfsslv1alpha1 "github.com/OpenSource-THG/cfssl-issuer/api/v1alpha1"
	"github.com/OpenSource-THG/cfssl-issuer/provisioners"
	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/go-logr/logr"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CertificateRequestReconciler reconciles a LocalCA object
type CertificateRequestReconciler struct {
	client.Client
	Log      logr.Logger
	Clock    clock.Clock
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=cert-manager.io,resources=certificaterequests,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificaterequests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

func (r *CertificateRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("certificaterequest", req.NamespacedName)

	// Fetch the CertificateRequest resource being reconciled
	cr := &cmapi.CertificateRequest{}
	if err := r.Client.Get(ctx, req.NamespacedName, cr); err != nil {
		log.Error(err, "failed to retrieve CertificateRequest resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check the CertificateRequest's issuerRef and if it does not match the
	// our group name, log a message at a debug level and stop processing.
	if cr.Spec.IssuerRef.Group != cfsslv1alpha1.GroupVersion.Group {
		log.V(4).Info("resource does not specify an issuerRef group name that we are responsible for", "group", cr.Spec.IssuerRef.Group)
		return ctrl.Result{}, nil
	}

	// Load the configured provisioner
	provisioner, err := LoadProvisioner(req, cr, log)
	if err != nil {
		_ = r.setStatus(ctx, cr, cmmetav1.ConditionFalse, cmapi.CertificateRequestReasonPending, "%s resource %s is not Ready", cr.Spec.IssuerRef.Kind, cr.Spec.IssuerRef.Name)
		return ctrl.Result{}, err
	}

	// Sign the SR and return the cert and ca
	signedPEM, ca, err := provisioner.Sign(cr.Spec.Request)
	if err != nil {
		log.Error(err, "failed to sign certificate request")
		return ctrl.Result{}, r.setStatus(ctx, cr, cmmetav1.ConditionFalse, cmapi.CertificateRequestReasonFailed, "Failed to sign certificate request: %v", err)
	}

	cr.Status.Certificate = signedPEM
	cr.Status.CA = ca

	return ctrl.Result{}, r.setStatus(ctx, cr, cmmetav1.ConditionTrue, cmapi.CertificateRequestReasonIssued, "Certificate Issued")
}

func LoadProvisioner(req ctrl.Request, cr *cmapi.CertificateRequest, log logr.Logger) (provisioners.Provisioner, error) {
	var p provisioners.Provisioner
	var ok bool

	issuerKey := types.NamespacedName{
		Name: cr.Spec.IssuerRef.Name,
	}

	kind := cr.Spec.IssuerRef.Kind

	switch kind {
	case "CfsslIssuer":
		issuerKey.Namespace = req.NamespacedName.Namespace
	case "CfsslClusterIssuer":
	default:
		return nil, fmt.Errorf("unknown kind %s", kind)
	}

	p, ok = provisioners.Load(issuerKey)
	if !ok {
		err := fmt.Errorf("provisioner %s not found", issuerKey)
		log.Error(err, fmt.Sprintf("failed to retrieve %s resource", kind), "name", issuerKey)
		return nil, err
	}

	return p, nil
}

func (r *CertificateRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cmapi.CertificateRequest{}).
		Complete(r)
}

func (r *CertificateRequestReconciler) setStatus(ctx context.Context, cr *cmapi.CertificateRequest, status cmmetav1.ConditionStatus, reason, message string, args ...interface{}) error {
	completeMessage := fmt.Sprintf(message, args...)
	r.setCondition(cr, cmapi.CertificateRequestConditionReady, status, reason, completeMessage)

	// Fire an Event to additionally inform users of the change
	eventType := core.EventTypeNormal
	if status == cmmetav1.ConditionFalse {
		eventType = core.EventTypeWarning
	}
	r.Recorder.Event(cr, eventType, reason, completeMessage)

	if err := r.Client.Update(ctx, cr); err != nil {
		r.Log.Error(err, "failed to update CertificateRequest")
		return err
	}

	if err := r.Client.Status().Update(ctx, cr); err != nil {
		r.Log.Error(err, "failed to update CertificateRequest status")
		return err
	}

	return nil
}

func (r *CertificateRequestReconciler) setCondition(cr *cmapi.CertificateRequest, conditionType cmapi.CertificateRequestConditionType, status cmmetav1.ConditionStatus, reason, message string) {
	now := meta.NewTime(r.Clock.Now())
	c := cmapi.CertificateRequestCondition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: &now,
	}

	// Search through existing conditions
	for idx, cond := range cr.Status.Conditions {
		// Skip unrelated conditions
		if cond.Type != conditionType {
			continue
		}

		// If this update doesn't contain a state transition, we don't update
		// the conditions LastTransitionTime to Now()
		if cond.Status == status {
			c.LastTransitionTime = cond.LastTransitionTime
		} else {
			r.Log.Info(fmt.Sprintf("Found status change for CertificateRequest %q condition %q: %q -> %q; setting lastTransitionTime to %v", cr.Name, conditionType, cond.Status, status, now.Time))
		}

		// Overwrite the existing condition
		cr.Status.Conditions[idx] = c
		return
	}

	// If we've not found an existing condition of this type, we simply insert
	// the new condition into the slice.
	cr.Status.Conditions = append(cr.Status.Conditions, c)
	r.Log.Info(fmt.Sprintf("Setting lastTransitionTime for CertificateRequest %q condition %q to %v", cr.Name, conditionType, now.Time))
}
