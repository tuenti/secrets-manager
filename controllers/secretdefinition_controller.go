/*
Copyright 2021.

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
	"reflect"
	"time"

	"github.com/go-logr/logr"
	smv1alpha1 "github.com/tuenti/secrets-manager/api/v1alpha1"
	"github.com/tuenti/secrets-manager/backend"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// https://golang.org/pkg/time/#pkg-constants
	timestampFormat = "2006-01-02T15.04.05Z"
	finalizerName   = "secret.finalizer." + smv1alpha1.Group
	managedByLabel  = "app.kubernetes.io/managed-by"
	lastUpdateLabel = smv1alpha1.Group + "/lastUpdateTime"
)

// SecretDefinitionReconciler reconciles a SecretDefinition object
type SecretDefinitionReconciler struct {
	client.Client
	Backend              backend.Client
	Log                  logr.Logger
	APIReader            client.Reader
	ReconciliationPeriod time.Duration
	ExcludeNamespaces    map[string]bool
	Scheme               *runtime.Scheme
}

// Annotations to skip when copying from a SecretDef to a Secret
var annotationsToSkip = make(map[string]bool)

// Helper functions to merge labels and annotations
type skipfn func(string) bool

func noSkip(_ string) bool {
	return false
}

func skipAnnotation(key string) bool {
	return annotationsToSkip[key]
}

func mergeMap(dst map[string]string, srcMap map[string]string, skipKey skipfn) {
	for k, v := range srcMap {
		if skipKey(k) {
			continue
		}
		dst[k] = v
	}
}

func getSecretFromSecretDefinition(sDef *smv1alpha1.SecretDefinition, data map[string][]byte) *corev1.Secret {
	objectMeta := getObjectMetaFromSecretDefinition(sDef)
	return &corev1.Secret{
		Type:       corev1.SecretType(sDef.Spec.Type),
		ObjectMeta: objectMeta,
		Data:       data,
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

// Ignore not found errors
func ignoreNotFoundError(err error) error {
	if errors.IsNotFound(err) {
		return nil
	}
	return err
}

// isNotMarkedForRemoval will determine if the SecretDefinition object has been marked to be deleted
func isNotMarkedForRemoval(sDef smv1alpha1.SecretDefinition) bool {
	return sDef.ObjectMeta.DeletionTimestamp.IsZero()
}

// getDesiredState reads the content from the Datasource for later comparison
func (r *SecretDefinitionReconciler) getDesiredState(keysMap map[string]smv1alpha1.DataSource) (map[string][]byte, error) {

	desiredState := make(map[string][]byte)
	var err error
	for k, v := range keysMap {
		bSecret, err := r.Backend.ReadSecret(v.Path, v.Key)
		if err != nil {
			r.Log.Error(err, "unable to read secret from backend", "path", v.Path, "key", v.Key)
			return nil, err
		}
		decoder, err := backend.NewDecoder(v.Encoding)
		if err != nil {
			r.Log.Error(err, "refusing to use encoding", "encoding", v.Encoding)
			return nil, err
		}
		desiredState[k], err = decoder.DecodeString(bSecret)
		if err != nil {
			r.Log.Error(err, "unable to decode data for secret", "encoding", v.Encoding, "path", v.Path, "key", v.Key)
			return nil, err
		}
	}
	return desiredState, err
}

// getCurrentState reads the content from the Kubernetes Secret API object for later comparison
func (r *SecretDefinitionReconciler) getCurrentState(ctx context.Context, namespace string, name string) (map[string][]byte, error) {
	// We don't read secrets from cache, as it's not the object we reconcile
	reader := r.APIReader
	data := make(map[string][]byte)
	secret := &corev1.Secret{}
	err := reader.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, secret)
	if err != nil {
		secretReadErrorsTotal.WithLabelValues(name, namespace).Inc()
		return data, err
	}
	data = secret.Data
	return data, err
}

// upsertSecret will create or update a secret
func (r *SecretDefinitionReconciler) upsertSecret(ctx context.Context, sDef *smv1alpha1.SecretDefinition, data map[string][]byte) error {
	secret := getSecretFromSecretDefinition(sDef, data)
	err := r.Create(ctx, secret)
	if errors.IsAlreadyExists(err) {
		err = r.Update(ctx, secret)
	}
	return err
}

// deleteSecret will delete a secret given its namespace and name
func (r *SecretDefinitionReconciler) deleteSecret(ctx context.Context, namespace string, name string) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
	return r.Delete(ctx, secret)
}

// shouldExclude will return true if the secretDefinition is in an excluded namespace
func (r *SecretDefinitionReconciler) shouldExclude(sDefNamespace string) bool {
	if len(r.ExcludeNamespaces) > 0 {
		return r.ExcludeNamespaces[sDefNamespace]
	}
	return false
}

// AddFinalizerIfNotPresent will check if finalizerName is the finalizers slice
func (r *SecretDefinitionReconciler) AddFinalizerIfNotPresent(ctx context.Context, sDef *smv1alpha1.SecretDefinition, finalizerName string) error {
	if !containsString(sDef.ObjectMeta.Finalizers, finalizerName) {
		sDef.ObjectMeta.Finalizers = append(sDef.ObjectMeta.Finalizers, finalizerName)
		return r.Update(ctx, sDef)
	}
	return nil
}

// Helper functions to manage corev1.Secret and smv1alpha1.SecretDefinition
func getObjectMetaFromSecretDefinition(sDef *smv1alpha1.SecretDefinition) metav1.ObjectMeta {
	labels := map[string]string{
		managedByLabel: "secrets-manager",
	}
	annotations := map[string]string{
		lastUpdateLabel: time.Now().Format(timestampFormat),
	}

	mergeMap(labels, sDef.Labels, noSkip)
	mergeMap(annotations, sDef.Annotations, skipAnnotation)

	return metav1.ObjectMeta{
		Namespace:   sDef.Namespace,
		Name:        sDef.Spec.Name,
		Labels:      labels,
		Annotations: annotations,
	}
}

//+kubebuilder:rbac:groups=secrets-manager.tuenti.io,resources=secretdefinitions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=secrets-manager.tuenti.io,resources=secretdefinitions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=secrets-manager.tuenti.io,resources=secretdefinitions/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SecretDefinition object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *SecretDefinitionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := r.Log.WithValues("secretdefinition", req.NamespacedName)
	sDef := &smv1alpha1.SecretDefinition{}

	err := r.Get(ctx, req.NamespacedName, sDef)
	if err != nil {
		log.Error(err, fmt.Sprintf("could not get SecretDefinition '%s'", req.NamespacedName))
		return ctrl.Result{}, ignoreNotFoundError(err)
	}

	secretName := sDef.Spec.Name
	secretNamespace := sDef.Namespace

	log = log.WithValues("secret", fmt.Sprintf("%s/%s", secretNamespace, secretName))

	if isNotMarkedForRemoval(*sDef) {

		err = r.AddFinalizerIfNotPresent(ctx, sDef, finalizerName)
		if err != nil {
			log.Error(err, "unable to update SecretDefinition finalizers", "finalizer", finalizerName)
			return ctrl.Result{}, err
		}

		if r.shouldExclude(sDef.Namespace) {
			log.Info("Secret definition in excluded namespace, ignoring", "excluded_namespaces", r.ExcludeNamespaces)
			return ctrl.Result{}, nil
		}
		// Get data from the secret source of truth
		desiredState, err := r.getDesiredState(sDef.Spec.KeysMap)

		if err != nil {
			log.Error(err, "unable to get desired state for secret")
			secretSyncErrorsTotal.WithLabelValues(secretNamespace, secretName).Inc()
			secretLastSyncStatus.WithLabelValues(secretNamespace, secretName).Set(0.0)
			return ctrl.Result{}, err
		}

		// Get the actual secret from Kubernetes
		currentState, err := r.getCurrentState(ctx, secretNamespace, secretName)

		if err != nil && !errors.IsNotFound(err) {
			log.Error(err, "unable to get current state of secret")
			secretSyncErrorsTotal.WithLabelValues(secretNamespace, secretName).Inc()
			secretLastSyncStatus.WithLabelValues(secretNamespace, secretName).Set(0.0)
			return ctrl.Result{}, ignoreNotFoundError(err)
		}

		eq := reflect.DeepEqual(desiredState, currentState)
		if !eq {
			log.Info("secret must be updated")
			if err := r.upsertSecret(ctx, sDef, desiredState); err != nil {
				log.Error(err, "unable to upsert secret")
				secretSyncErrorsTotal.WithLabelValues(secretNamespace, secretName).Inc()
				secretLastSyncStatus.WithLabelValues(secretNamespace, secretName).Set(0.0)
				return ctrl.Result{}, err
			}
			log.Info("secret updated")
		}
		secretLastSyncStatus.WithLabelValues(secretNamespace, secretName).Set(1.0)
		return ctrl.Result{RequeueAfter: r.ReconciliationPeriod}, nil

	} else {
		// SecretDefinition has been marked for deletion and contains finalizer
		if controllerutil.ContainsFinalizer(sDef, finalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err = r.deleteSecret(ctx, secretNamespace, secretName); err != nil && !errors.IsNotFound(err) {
				log.Error(err, "Unable to delete secret")
				return ctrl.Result{}, ignoreNotFoundError(err)
			}
			log.Info("secret deleted successfully")

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(sDef, finalizerName)
			if err := r.Update(ctx, sDef); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretDefinitionReconciler) SetupWithManager(mgr ctrl.Manager, name string) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&smv1alpha1.SecretDefinition{}).
		Named(name).
		Complete(r)
}

func init() {
	// last-applied-configuration should not be copied from the SecretDef to the Secret
	annotationsToSkip[corev1.LastAppliedConfigAnnotation] = true
}
