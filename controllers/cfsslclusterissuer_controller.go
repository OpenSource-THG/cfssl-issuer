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

	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	certmanagerv1beta1 "github.com/OpenSource-THG/cfssl-issuer/api/v1beta1"
	"github.com/OpenSource-THG/cfssl-issuer/provisioners"
)

const errorReason = "Error"
const errorValidation = "Validation"
const initProvisionerFailure = "failed to initialize provisioner"

// CfsslClusterIssuerReconciler reconciles a CfsslClusterIssuer object
type CfsslClusterIssuerReconciler struct {
	client.Client
	Log      logr.Logger
	Clock    clock.Clock
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=certmanager.thg.io,resources=cfsslclusterissuers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=certmanager.thg.io,resources=cfsslclusterissuers/status,verbs=get;update;patch

func (r *CfsslClusterIssuerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("cfsslclusterissuer", req.NamespacedName)

	// Fetch the Cfssl resource being synced
	cfssl := &certmanagerv1beta1.CfsslClusterIssuer{}
	if err := r.Client.Get(ctx, req.NamespacedName, cfssl); err != nil {
		log.Error(err, "failed to retrieve Cfssl resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	statusReconciler := newCfsslClusterStatusReconciler(r, cfssl, log)
	if err := validateCfsslIssuerSpec(cfssl.Spec); err != nil {
		log.Error(err, "failed to validate CfsslClusterIssuer resource")
		_ = statusReconciler.Update(ctx, certmanagerv1beta1.ConditionFalse, errorValidation, "Failed to validate resource: %v", err)
		return ctrl.Result{}, err
	}

	p, err := provisioners.New(cfssl.Spec)
	if err != nil {
		log.Error(err, initProvisionerFailure)
		_ = statusReconciler.Update(ctx, certmanagerv1beta1.ConditionFalse, errorReason, initProvisionerFailure)
		return ctrl.Result{}, err
	}

	provisioners.Store(req.NamespacedName, p)

	return ctrl.Result{}, statusReconciler.Update(
		ctx, certmanagerv1beta1.ConditionTrue, "Verified", "CfsslClusterIssuer verified and ready to sign certificates")
}

func (r *CfsslClusterIssuerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&certmanagerv1beta1.CfsslClusterIssuer{}).
		Complete(r)
}
