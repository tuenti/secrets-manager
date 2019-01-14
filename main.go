package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/tuenti/secrets-manager/backend"
	k8s "github.com/tuenti/secrets-manager/kubernetes"
	"github.com/tuenti/secrets-manager/secrets-manager"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// To be filled from build ldflags
var version string

func newK8sClientSet() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientSet, nil
}

func main() {
	var logger *log.Logger
	var wg sync.WaitGroup

	backendCfg := backend.Config{}
	secretsManagerCfg := secretsmanager.Config{}
	selectedBackend := flag.String("backend", "vault", "Selected backend. Only vault supported")
	logLevel := flag.String("log.level", "warn", "Minimum log level")
	logFormat := flag.String("log.format", "text", "Log format, one of text or json")
	versionFlag := flag.Bool("version", false, "Display Secret Manager version")
	addr := flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")

	flag.StringVar(&secretsManagerCfg.ConfigMap, "config.config-map", "secrets-manager-config", "Name of the config Map with Secrets Manager settings (format: [<namespace>/]<name>) ")
	flag.DurationVar(&secretsManagerCfg.BackendScrapeInterval, "config.backend-scrape-interval", 15*time.Second, "Scraping secrets from backend interval")
	flag.DurationVar(&secretsManagerCfg.ConfigMapRefreshInterval, "config.configmap-refresh-interval", 15*time.Second, "ConfigMap refresh interval")

	flag.StringVar(&backendCfg.VaultURL, "vault.url", "https://127.0.0.1:8200", "Vault address. VAULT_ADDR environment would take precedence.")
	flag.StringVar(&backendCfg.VaultToken, "vault.token", "", "Vault token. VAULT_TOKEN environment would take precedence.")
	flag.Int64Var(&backendCfg.VaultMaxTokenTTL, "vault.max-token-ttl", 300, "Max seconds to consider a token expired.")
	flag.DurationVar(&backendCfg.VaultTokenPollingPeriod, "vault.token-polling-period", 15*time.Second, "Polling interval to check token expiration time.")
	flag.IntVar(&backendCfg.VaultRenewTTLIncrement, "vault.renew-ttl-increment", 600, "TTL time for renewed token.")
	flag.StringVar(&backendCfg.VaultEngine, "vault.engine", "kv2", "Vault secret engine. Only KV version 1 and 2 supported")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Secrets Manager %s\n", version)
		os.Exit(0)
	}

	logger = log.New()

	switch *logLevel {
	case "info":
		logger.SetLevel(log.InfoLevel)
	case "err":
		logger.SetLevel(log.ErrorLevel)
	case "debug":
		logger.SetLevel(log.DebugLevel)
	default:
		logger.SetLevel(log.WarnLevel)
	}

	switch *logFormat {
	case "json":
		logger.Formatter = &log.JSONFormatter{}
	default:
		logger.Formatter = &log.TextFormatter{}
	}

	logger.SetOutput(os.Stdout)

	if os.Getenv("VAULT_ADDR") != "" {
		backendCfg.VaultURL = os.Getenv("VAULT_ADDR")
	}

	if os.Getenv("VAULT_TOKEN") != "" {
		backendCfg.VaultToken = os.Getenv("VAULT_TOKEN")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	backendClient, err := backend.NewBackendClient(ctx, *selectedBackend, logger, backendCfg)
	if err != nil {
		logger.Errorf("could not build backend client: %v", err)
		os.Exit(1)
	}

	clientSet, err := newK8sClientSet()

	if err != nil {
		logger.Errorf("could not build k8s client: %v", err)
		os.Exit(1)
	}

	kubernetes := k8s.New(clientSet, logger)
	secretsManager, err := secretsmanager.New(ctx, secretsManagerCfg, kubernetes, *backendClient, logger)

	if err != nil {
		logger.Errorf("could not init Secret Manager: %v", err)
		os.Exit(1)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	wg.Add(1)
	go secretsManager.Start(ctx)

	srv := startHttpServer(*addr, logger)

	for {
		select {
		case <-sigc:
			shutdownHttpServer(srv, logger)
			cancel()
			break
		}
		break
	}
	wg.Wait()
}

func shutdownHttpServer(srv *http.Server, logger *log.Logger) {
	logger.Infof("[main] Stopping HTTP server")

    if err := srv.Shutdown(nil); err != nil {
        logger.Errorf("ListenAndServe(): %s", err)
    } else {
		logger.Infof("[main] Stopped HTTP server")
	}
}

func startHttpServer(addr string, logger *log.Logger) *http.Server {
    srv := &http.Server{Addr: addr}

    http.Handle("/metrics", promhttp.Handler())

    go func() {
		logger.Infof("Starting HTTP server listening on %v", addr)
        // returns ErrServerClosed on graceful close
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            logger.Errorf("[main] Unexpected error in HTTP server: %s", err)
        }
    }()

    return srv
}