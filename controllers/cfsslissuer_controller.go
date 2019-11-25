/*

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

	certmanagerv1beta1 "github.com/OpenSource-THG/cfssl-issuer/api/v1beta1"
	"github.com/OpenSource-THG/cfssl-issuer/provisioners"
	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CfsslIssuerReconciler reconciles a CfsslIssuer object
type CfsslIssuerReconciler struct {
	client.Client
	Log      logr.Logger
	Clock    clock.Clock
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=certmanager.thg.io,resources=cfsslissuers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=certmanager.thg.io,resources=cfsslissuers/status,verbs=get;update;patch

// Reconcile reconciles a given CfsslIssuer resource
func (r *CfsslIssuerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("cfsslissuer", req.NamespacedName)

	// Fetch the Cfssl resource being synced
	cfssl := &certmanagerv1beta1.CfsslIssuer{}
	if err := r.Client.Get(ctx, req.NamespacedName, cfssl); err != nil {
		log.Error(err, "failed to retrieve Cfssl resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	finalizer := "cfsslissuer.finalizers.certmanager.thg.io"

	// Check if deletion timestamp is set; if false object is under deletion
	if cfssl.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(cfssl.ObjectMeta.Finalizers, finalizer) {
			cfssl.ObjectMeta.Finalizers = append(cfssl.ObjectMeta.Finalizers, finalizer)
			if err := r.Update(ctx, cfssl); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if containsString(cfssl.ObjectMeta.Finalizers, finalizer) {
			// Remove issuer from provisioners
			provisioners.Remove(req.NamespacedName)
			cfssl.ObjectMeta.Finalizers = removeString(cfssl.ObjectMeta.Finalizers, finalizer)
			if err := r.Update(ctx, cfssl); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	statusReconciler := newCfsslStatusReconciler(r, cfssl, log)
	if err := validateCfsslIssuerSpec(cfssl.Spec); err != nil {
		log.Error(err, "failed to validate CfsslIssuer resource")
		_ = statusReconciler.Update(ctx, certmanagerv1beta1.ConditionFalse, "Validation", "Failed to validate resource: %v", err)
		return ctrl.Result{}, err
	}

	p, err := provisioners.New(cfssl.Spec)
	if err != nil {
		log.Error(err, "failed to initialize provisioner")
		_ = statusReconciler.Update(ctx, certmanagerv1beta1.ConditionFalse, "Error", "failed to initialize provisioner")
		return ctrl.Result{}, err
	}

	provisioners.Store(req.NamespacedName, p)

	return ctrl.Result{}, statusReconciler.Update(ctx, certmanagerv1beta1.ConditionTrue, "Verified", "CfsslIssuer verified and ready to sign certificates")
}

// SetupWithManager registers CfsslIssuerReconciler with the given manager
func (r *CfsslIssuerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&certmanagerv1beta1.CfsslIssuer{}).
		Complete(r)
}

func validateCfsslIssuerSpec(c *certmanagerv1beta1.CfsslIssuerSpec) error {
	switch {
	case c.URL == "":
		return fmt.Errorf("spec.url cannot be empty")
	case c.CABundle == nil:
		return fmt.Errorf("spec.caBundle cannot be empty")
	default:
		return nil
	}
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
