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

	"github.com/go-logr/logr"
	apiutil "github.com/jetstack/cert-manager/pkg/api/util"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	cmmetav1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	cfsslv1beta1 "github.com/opensource-thg/cfssl-issuer/api/v1beta1"
	"github.com/opensource-thg/cfssl-issuer/provisioners"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CertificateRequestReconciler reconciles a LocalCA object
type CertificateRequestReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=cert-manager.io,resources=certificaterequests,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificaterequests/status,verbs=get;update;patch

func (r *CertificateRequestReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("certificaterequest", req.NamespacedName)

	// Fetch the CertificateRequest resource being reconciled
	cr := &cmapi.CertificateRequest{}
	if err := r.Client.Get(ctx, req.NamespacedName, cr); err != nil {
		log.Error(err, "failed to retrieve CertificateRequest resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check the CertificateRequest's issuerRef and if it does not match the
	// exampleapi group name, log a message at a debug level and stop processing.
	if cr.Spec.IssuerRef.Group != cfsslv1beta1.GroupVersion.Group {
		log.V(4).Info("resource does not specify an issuerRef group name that we are responsible for", "group", cr.Spec.IssuerRef.Group)
		return ctrl.Result{}, nil
	}

	issNamespaceName := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      cr.Spec.IssuerRef.Name,
	}
	provisioner, ok := provisioners.Load(issNamespaceName)
	if !ok {
		err := fmt.Errorf("provisioner %s not found", issNamespaceName)
		log.Error(err, "failed to retrieve CfsslIssuer resource", "namespace", req.Namespace, "name", cr.Spec.IssuerRef.Name)
		r.setStatus(ctx, cr, cmmetav1.ConditionFalse, cmapi.CertificateRequestReasonPending, "CfsslIssuer resource %s is not Ready", issNamespaceName)
		return ctrl.Result{}, err
	}

	signedPEM, ca, err := provisioner.Sign(ctx, cr)
	if err != nil {
		log.Error(err, "failed to sign certificate request")
		return ctrl.Result{}, r.setStatus(ctx, cr, cmmetav1.ConditionFalse, cmapi.CertificateRequestReasonFailed, "Failed to sign certificate request: %v", err)
	}

	cr.Status.Certificate = signedPEM
	cr.Status.CA = ca

	return ctrl.Result{}, r.setStatus(ctx, cr, cmmetav1.ConditionTrue, cmapi.CertificateRequestReasonIssued, "Certificate Issued")
}

func (r *CertificateRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cmapi.CertificateRequest{}).
		Complete(r)
}

func (r *CertificateRequestReconciler) setStatus(ctx context.Context, cr *cmapi.CertificateRequest, status cmmetav1.ConditionStatus, reason, message string, args ...interface{}) error {
	completeMessage := fmt.Sprintf(message, args...)
	apiutil.SetCertificateRequestCondition(cr, cmapi.CertificateRequestConditionReady, status, reason, completeMessage)

	// Fire an Event to additionally inform users of the change
	eventType := core.EventTypeNormal
	if status == cmmetav1.ConditionFalse {
		eventType = core.EventTypeWarning
	}
	r.Recorder.Event(cr, eventType, reason, completeMessage)

	return r.Client.Update(ctx, cr)
}
