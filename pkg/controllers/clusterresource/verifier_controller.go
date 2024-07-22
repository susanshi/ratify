/*
Copyright The Ratify Authors.
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

package clusterresource

import (
	"context"
	"fmt"

	configv1beta1 "github.com/ratify-project/ratify/api/v1beta1"
	"github.com/ratify-project/ratify/internal/constants"
	"github.com/ratify-project/ratify/pkg/controllers"

	cutils "github.com/ratify-project/ratify/pkg/controllers/utils"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// VerifierReconciler reconciles a Verifier object
type VerifierReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=verifiers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=verifiers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=verifiers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Verifier object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *VerifierReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	verifierLogger := logrus.WithContext(ctx)

	var verifier configv1beta1.Verifier
	var resource = req.Name

	verifierLogger.Infof("reconciling verifier '%v'", resource)

	if err := r.Get(ctx, req.NamespacedName, &verifier); err != nil {
		if apierrors.IsNotFound(err) {
			verifierLogger.Infof("delete event detected, removing verifier %v", resource)
			controllers.NamespacedVerifiers.DeleteVerifier(constants.EmptyNamespace, resource)
		} else {
			verifierLogger.Error(err, "unable to fetch verifier")
		}

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := verifierAddOrReplace(verifier.Spec, resource); err != nil {
		verifierLogger.Error(err, "unable to create verifier from verifier crd")
		writeVerifierStatus(ctx, r, &verifier, verifierLogger, false, err.Error())
		return ctrl.Result{}, err
	}

	writeVerifierStatus(ctx, r, &verifier, verifierLogger, true, "")

	// returning empty result and no error to indicate we’ve successfully reconciled this object
	return ctrl.Result{}, nil
}

// creates a verifier reference from CRD spec and add store to map
func verifierAddOrReplace(spec configv1beta1.VerifierSpec, objectName string) error {
	verifierConfig, err := cutils.SpecToVerifierConfig(spec.Parameters.Raw, objectName, spec.Name, spec.ArtifactTypes, spec.Source)
	if err != nil {
		logrus.Error(err, "unable to convert crd specification to verifier config")
		return fmt.Errorf("unable to convert crd specification to verifier config, err: %w", err)
	}

	return cutils.UpsertVerifier(spec.Version, spec.Address, constants.EmptyNamespace, objectName, verifierConfig)
}

// SetupWithManager sets up the controller with the Manager.
func (r *VerifierReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1beta1.Verifier{}).
		Complete(r)
}

func writeVerifierStatus(ctx context.Context, r client.StatusClient, verifier *configv1beta1.Verifier, logger *logrus.Entry, isSuccess bool, errorString string) {
	if isSuccess {
		verifier.Status.IsSuccess = true
		verifier.Status.Error = ""
		verifier.Status.BriefError = ""
	} else {
		verifier.Status.IsSuccess = false
		verifier.Status.Error = errorString
		if len(errorString) > constants.MaxBriefErrLength {
			verifier.Status.BriefError = fmt.Sprintf("%s...", errorString[:constants.MaxBriefErrLength])
		}
	}

	if statusErr := r.Status().Update(ctx, verifier); statusErr != nil {
		logger.Error(statusErr, ",unable to update verifier status")
	}
}
