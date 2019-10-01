package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	secretsmanagerv1alpha1 "github.com/tuenti/secrets-manager/api/v1alpha1"
	"github.com/tuenti/secrets-manager/backend"
	"github.com/tuenti/secrets-manager/controllers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	corev1.AddToScheme(scheme)
	secretsmanagerv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

// To be filled from build ldflags
var version string

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var enableDebugLog bool
	var versionFlag bool
	var reconcilePeriod time.Duration
	var selectedBackend string
	var watchNamespaces string
	var excludeNamespaces string

	backendCfg := backend.Config{}

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&selectedBackend, "backend", "vault", "Selected backend. Only vault supported")
	flag.BoolVar(&enableDebugLog, "enable-debug-log", false, "Enable this to get more logs verbosity and debug messages.")
	flag.BoolVar(&versionFlag, "version", false, "Display Secret Manager version")
	flag.DurationVar(&reconcilePeriod, "reconcile-period", 5*time.Second, "How often the controller will re-queue secretdefinition events")
	flag.DurationVar(&backendCfg.BackendTimeout, "config.backend-timeout", 5*time.Second, "Backend connection timeout")
	flag.StringVar(&backendCfg.VaultURL, "vault.url", "https://127.0.0.1:8200", "Vault address. VAULT_ADDR environment would take precedence.")
	flag.StringVar(&backendCfg.VaultRoleID, "vault.role-id", "", "Vault approle role id. VAULT_ROLE_ID environment would take precedence.")
	flag.StringVar(&backendCfg.VaultSecretID, "vault.secret-id", "", "Vault approle secret id. VAULT_SECRET_ID environment would take precedence.")
	flag.Int64Var(&backendCfg.VaultMaxTokenTTL, "vault.max-token-ttl", 300, "Max seconds to consider a token expired.")
	flag.DurationVar(&backendCfg.VaultTokenPollingPeriod, "vault.token-polling-period", 15*time.Second, "Polling interval to check token expiration time.")
	flag.IntVar(&backendCfg.VaultRenewTTLIncrement, "vault.renew-ttl-increment", 600, "TTL time for renewed token.")
	flag.StringVar(&backendCfg.VaultEngine, "vault.engine", "kv2", "Vault secret engine. Only KV version 1 and 2 supported")
	flag.StringVar(&watchNamespaces, "watch-namespaces", "", "Comma separated list of namespaces that secrets-manager will watch for SecretDefinitions. By default all namespaces are watched.")
	flag.StringVar(&excludeNamespaces, "exclude-namespaces", "", "Comma separated list of namespaces that secrets-manager will not watch for SecretDefinitions. By default all namespaces are watched.")
	flag.Parse()

	if versionFlag {
		fmt.Printf("Secrets Manager %s\n", version)
		os.Exit(0)
	}

	logger := zap.Logger(enableDebugLog).WithName("backend")

	if os.Getenv("VAULT_ADDR") != "" {
		backendCfg.VaultURL = os.Getenv("VAULT_ADDR")
	}

	if os.Getenv("VAULT_ROLE_ID") != "" {
		backendCfg.VaultRoleID = os.Getenv("VAULT_ROLE_ID")
	}

	if os.Getenv("VAULT_SECRET_ID") != "" {
		backendCfg.VaultSecretID = os.Getenv("VAULT_SECRET_ID")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	backendClient, err := backend.NewBackendClient(ctx, selectedBackend, logger, backendCfg)
	if err != nil {
		logger.Error(err, "could not build backend client")
		os.Exit(1)
	}

	ctrl.SetLogger(zap.Logger(enableDebugLog))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	nsSlice := func(ns string) []string {
		trimmed := strings.Trim(strings.TrimSpace(ns), "\"")
		return strings.Split(trimmed, ",")
	}

	watchNs := make(map[string]bool)
	if len(watchNamespaces) > 0 {
		for _, ns := range nsSlice(watchNamespaces) {
			watchNs[ns] = true
		}
	}
	if len(excludeNamespaces) > 0 {
		for _, ns := range nsSlice(excludeNamespaces) {
			watchNs[ns] = false
		}
	}

	err = (&controllers.SecretDefinitionReconciler{
		Backend:              *backendClient,
		Client:               mgr.GetClient(),
		APIReader:            mgr.GetAPIReader(),
		Log:                  ctrl.Log.WithName("controllers").WithName("SecretDefinition"),
		Ctx:                  ctx,
		ReconciliationPeriod: reconcilePeriod,
		WatchNamespaces:      watchNs,
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SecretDefinition")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
