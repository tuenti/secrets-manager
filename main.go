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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	//logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	secretsmanagerv1alpha1 "github.com/tuenti/secrets-manager/api/v1alpha1"
	"github.com/tuenti/secrets-manager/backend"
	"github.com/tuenti/secrets-manager/controllers"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	corev1.AddToScheme(scheme)
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(secretsmanagerv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

// To be filled from build ldflags
var version string

func main() {
	var metricsAddr string
	var controllerName string
	var enableLeaderElection bool
	var enableDebugLog bool
	var versionFlag bool
	var reconcilePeriod time.Duration
	var selectedBackend string
	var watchNamespaces string
	var excludeNamespaces string
	var mgr ctrl.Manager
	var namespaceList []string

	backendCfg := backend.Config{}

	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&controllerName, "controller-name", "SecretDefinition", "If running secrets manager in multiple namespaces, set the controller name to something unique avoid 'duplicate metrics collector registration attempted' errors.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&selectedBackend, "backend", "vault", "Selected backend. One of vault or azure-kv")
	flag.BoolVar(&enableDebugLog, "enable-debug-log", false, "Enable this to get more logs verbosity and debug messages.")
	flag.BoolVar(&versionFlag, "version", false, "Display Secret Manager version")
	flag.DurationVar(&reconcilePeriod, "reconcile-period", 5*time.Second, "How often the controller will re-queue secretdefinition events")
	flag.DurationVar(&backendCfg.BackendTimeout, "config.backend-timeout", 5*time.Second, "Backend connection timeout")
	flag.StringVar(&backendCfg.VaultURL, "vault.url", "https://127.0.0.1:8200", "Vault address. VAULT_ADDR environment would take precedence.")
	flag.StringVar(&backendCfg.VaultAuthMethod, "vault.auth-method", "approle", "Vault authentication method. Supported: approle, kubernetes.")
	flag.StringVar(&backendCfg.VaultRoleID, "vault.role-id", "", "Vault approle role id. VAULT_ROLE_ID environment would take precedence.")
	flag.StringVar(&backendCfg.VaultSecretID, "vault.secret-id", "", "Vault approle secret id. VAULT_SECRET_ID environment would take precedence.")
	flag.StringVar(&backendCfg.VaultKubernetesRole, "vault.kubernetes-role", "", "Vault kubernetes role name.")
	flag.Int64Var(&backendCfg.VaultMaxTokenTTL, "vault.max-token-ttl", 300, "Max seconds to consider a token expired.")
	flag.DurationVar(&backendCfg.VaultTokenPollingPeriod, "vault.token-polling-period", 15*time.Second, "Polling interval to check token expiration time.")
	flag.IntVar(&backendCfg.VaultRenewTTLIncrement, "vault.renew-ttl-increment", 600, "TTL time for renewed token.")
	flag.StringVar(&backendCfg.VaultEngine, "vault.engine", "kv2", "Vault secret engine. Only KV version 1 and 2 supported")
	flag.StringVar(&backendCfg.VaultApprolePath, "vault.approle-path", "approle", "Vault approle login path")
	flag.StringVar(&backendCfg.VaultKubernetesPath, "vault.kubernetes-path", "kubernetes", "Vault kubernetes login path")
	flag.StringVar(&backendCfg.AzureKVName, "azure-kv.name", "", "Azure KeyVault name. AZURE_KV_NAME environment would take precedence")
	flag.StringVar(&backendCfg.AzureKVTenantID, "azure-kv.tenant-id", "", "Azure KeyVault Tenant ID. AZURE_TENANT_ID environment would take precedence")
	flag.StringVar(&backendCfg.AzureKVClientID, "azure-kv.client-id", "", "Azure KeyVault ClientID used to authenticate. AZURE_CLIENT_ID environment would take precedence")
	flag.StringVar(&backendCfg.AzureKVClientSecret, "azure-kv.client-secret", "", "Azure KeyVault Client Secret used to authenticate. AZURE_CLIENT_SECRET environment would take precedence")
	flag.StringVar(&backendCfg.AzureKVManagedClientID, "azure-kv.managed-client-id", "", "Azure KeyVault Client ID used to authenticate using managed identity. AZURE_MANAGED_CLIENT_ID environment would take precedence")
	flag.StringVar(&backendCfg.AzureKVManagedResourceID, "azure-kv.managed-resource-id", "", "Azure KeyVault Resource ID used to authenticate using managed identity. AZURE_MANAGED_RESOURCE_ID environment would take precedence")
	flag.StringVar(&watchNamespaces, "watch-namespaces", "", "Comma separated list of namespaces that secrets-manager will watch for SecretDefinitions. By default all namespaces are watched.")
	flag.StringVar(&excludeNamespaces, "exclude-namespaces", "", "Comma separated list of namespaces that secrets-manager will not watch for SecretDefinitions. By default all namespaces are watched.")

	//New
	opts := zap.Options{
		Development: enableDebugLog,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	backendLog := ctrl.Log.WithName("backend")

	if versionFlag {
		fmt.Printf("Secrets Manager %s\n", version)
		os.Exit(0)
	}

	if os.Getenv("VAULT_ADDR") != "" {
		backendCfg.VaultURL = os.Getenv("VAULT_ADDR")
	}

	if os.Getenv("VAULT_ROLE_ID") != "" {
		backendCfg.VaultRoleID = os.Getenv("VAULT_ROLE_ID")
	}

	if os.Getenv("VAULT_SECRET_ID") != "" {
		backendCfg.VaultSecretID = os.Getenv("VAULT_SECRET_ID")
	}

	if os.Getenv("AZURE_KV_NAME") != "" {
		backendCfg.AzureKVName = os.Getenv("AZURE_KV_NAME")
	}

	if os.Getenv("AZURE_MANAGED_CLIENT_ID") != "" {
		backendCfg.AzureKVManagedClientID = os.Getenv("AZURE_MANAGED_CLIENT_ID")
	}

	if os.Getenv("AZURE_MANAGED_RESOURCE_ID") != "" {
		backendCfg.AzureKVManagedResourceID = os.Getenv("AZURE_MANAGED_RESOURCE_ID")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	backendClient, err := backend.NewBackendClient(ctx, selectedBackend, backendLog, backendCfg)
	if err != nil {
		setupLog.Error(err, "could not build backend client")
		os.Exit(1)
	}

	nsSlice := func(ns string) []string {
		trimmed := strings.Trim(strings.TrimSpace(ns), "\"")
		return strings.Split(trimmed, ",")
	}

	excludeNs := make(map[string]bool)
	if len(excludeNamespaces) > 0 {
		for _, ns := range nsSlice(excludeNamespaces) {
			excludeNs[ns] = true
		}
	}

	if len(strings.TrimSpace(watchNamespaces)) > 0 {
		setupLog.Info("setting restricted namespace list for controller")
		namespaceList = nsSlice(watchNamespaces)
		setupLog.Info("watching namespaces: " + watchNamespaces)
		mgr, err = ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
			Scheme:                 scheme,
			MetricsBindAddress:     metricsAddr,
			HealthProbeBindAddress: probeAddr,
			LeaderElection:         enableLeaderElection,
			LeaderElectionID:       "5ac9a181.secrets-manager.tuenti.io",
			NewCache:               cache.MultiNamespacedCacheBuilder(namespaceList),
		})
		if err != nil {
			setupLog.Error(err, "unable to start manager")
			os.Exit(1)
		}
	} else {
		mgr, err = ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
			Scheme:                 scheme,
			MetricsBindAddress:     metricsAddr,
			HealthProbeBindAddress: probeAddr,
			LeaderElection:         enableLeaderElection,
			LeaderElectionID:       "5ac9a181.secrets-manager.tuenti.io",
		})
		if err != nil {
			setupLog.Error(err, "unable to start manager")
			os.Exit(1)
		}
	}

	if err = (&controllers.SecretDefinitionReconciler{
		Client:               mgr.GetClient(),
		Backend:              *backendClient,
		APIReader:            mgr.GetAPIReader(),
		Log:                  ctrl.Log.WithName("controllers").WithName("SecretDefinition"),
		ReconciliationPeriod: reconcilePeriod,
		ExcludeNamespaces:    excludeNs,
	}).SetupWithManager(mgr, controllerName); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SecretDefinition")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
