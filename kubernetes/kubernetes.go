package kubernetes

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	log "github.com/sirupsen/logrus"
	smerrors "github.com/tuenti/secrets-manager/errors"
)

var logger *log.Logger

// Secret represents a K8s secret object
type Secret struct {
	Name      string
	Namespace string
	Data      map[string][]byte
	Type      string
	Labels    map[string]string
}

// Client provides a facade on the K8s API
type Client interface {
	UpsertSecret(secret *Secret) error
	ReadSecret(namespace string, name string) (map[string][]byte, error)
	ReadConfigMap(name string, namespace string, key string) (string, error)
}

type client struct {
	client kubernetes.Interface
}

// New creates a K8s client
func New(clientSet kubernetes.Interface, l *log.Logger) Client {
	k := &client{
		client: clientSet,
	}
	logger = l
	return k
}

func (k *client) UpsertSecret(secret *Secret) error {
	k8sSecret := &corev1.Secret{
		Type: corev1.SecretType(secret.Type),
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Name,
			Labels:    secret.Labels,
			Namespace: secret.Namespace,
		},
		Data: secret.Data,
	}
	_, err := k.client.CoreV1().Secrets(secret.Namespace).Get(secret.Name, metav1.GetOptions{})

	if err != nil && errors.IsNotFound(err) {
		logger.Debugf("creating secret '%s/%s'", secret.Namespace, secret.Name)
		_, err = k.client.CoreV1().Secrets(secret.Namespace).Create(k8sSecret)
	} else {
		logger.Debugf("updating secret '%s/%s'", secret.Namespace, secret.Name)
		_, err = k.client.CoreV1().Secrets(secret.Namespace).Update(k8sSecret)
	}
	if err != nil {
		secretUpdateErrorsTotal.WithLabelValues(secret.Name, secret.Namespace).Inc()
	}
	return err
}

// ReadSecret returns a particular key in Kubernetes secrets object
func (k *client) ReadSecret(namespace string, name string) (map[string][]byte, error) {
	data := make(map[string][]byte)
	secret, err := k.client.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			secretReadErrorsTotal.WithLabelValues(name, namespace).Inc()
			return data, &smerrors.K8sSecretNotFoundError{ErrType: smerrors.K8sSecretNotFoundErrorType, Name: name, Namespace: namespace}
		}
		secretReadErrorsTotal.WithLabelValues(name, namespace).Inc()
		return data, err
	}

	data = secret.Data
	return data, err
}

func (k *client) ReadConfigMap(name string, namespace string, key string) (string, error) {
	configMap, err := k.client.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	configString := configMap.Data[key]
	return configString, nil
}
