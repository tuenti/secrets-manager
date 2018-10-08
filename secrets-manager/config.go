package secretsmanager

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v2"
)

// Config holds the general global Secret manager config
type Config struct {
	ConfigMapRefreshInterval time.Duration
	BackendScrapeInterval    time.Duration
	ConfigMap                string
}

// SecretDefinitions is a list of SecretDefinitions
type SecretDefinitions []SecretDefinition

// SecretDefinition defines how to generate a secret in K8s from remote secrets backends
type SecretDefinition struct {
	// Name for the secret in K8s
	Name string `yaml:"name"`
	// Namespaces is the list of namespaces where the secret is going to be created
	Namespaces []string `yaml:"namespaces"`
	// Type is the type of K8s Secret ("Opaque", "kubernetes.io/tls", ...)
	Type string `yaml:"type"`
	// Data is a dictionary which keys are the name of each entry in the K8s Secret data and the value is
	// the Datasource (from backend) for that entry
	Data map[string]Datasource `yaml:"data"` //optional?
}

// Datasource represents a reference to a secret in a backend (source of truth)
type Datasource struct {
	// Path to a secret in a secret backend
	Path string `yaml:"path"`
	// Key in the secret in the backend
	Key string `yaml:"key"`
	// Encoding type for the secret. Only base64 supported. Optional
	Encoding string `yaml:"encoding,omitempty"`
}

func parseSecretDefsFromYaml(configText string) (SecretDefinitions, error) {
	secretDefs := new([]SecretDefinition)

	err := yaml.Unmarshal([]byte(configText), secretDefs)
	if err != nil {
		fmt.Printf("error: could'n unmarshal yaml %v\n", err)
		return nil, err
	}
	return *secretDefs, nil
}
