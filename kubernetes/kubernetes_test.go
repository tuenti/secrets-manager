package kubernetes_test

import (
	"fmt"
	"os"

	"github.com/tuenti/secrets-manager/errors"
	"github.com/tuenti/secrets-manager/kubernetes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"testing"

	"github.com/tuenti/secrets-manager/testutils"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"

	log "github.com/sirupsen/logrus"
)

func TestUpsertSecretDoesNotExist(t *testing.T) {
	client := fake.NewSimpleClientset()

	k8s := kubernetes.New(client, log.New())

	k8sSecret := testutils.NewFakeSecret("ns", "secret-test")

	k8s.UpsertSecret(k8sSecret)

	secret, _ := client.CoreV1().Secrets("ns").Get("secret-test", metav1.GetOptions{})

	assert.NotNil(t, secret)
}

func TestUpsertSecretAlreadyExists(t *testing.T) {
	// Create the fake client.
	client := fake.NewSimpleClientset()

	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	logger.Out = os.Stderr

	k8s := kubernetes.New(client, logger)

	k8sSecret := testutils.NewFakeSecret("ns", "secret-test")

	// Upsert twice, second must be an update
	k8s.UpsertSecret(k8sSecret)
	k8s.UpsertSecret(k8sSecret)

	actions := client.Actions()
	lastAction := actions[len(actions)-1]
	assert.Implements(t, (*clientgotesting.UpdateAction)(nil), lastAction, "Last action must be UpdateAction")
}

func TestReadConfigMap(t *testing.T) {

	// Create the fake client.
	client := fake.NewSimpleClientset()

	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	logger.Out = os.Stderr

	k8s := kubernetes.New(client, logger)

	client.CoreV1().ConfigMaps("default").Create(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cm",
		},
		Data: map[string]string{
			"config": "fake-content",
		},
	})

	configMapContent, err := k8s.ReadConfigMap("cm", "default", "config")
	assert.Nil(t, err)
	assert.Equal(t, "fake-content", configMapContent)
}

func TestReadSecret(t *testing.T) {
	client := fake.NewSimpleClientset()

	k8s := kubernetes.New(client, log.New())

	data := "some-value"

	_, err := client.CoreV1().Secrets("ns").Create(&corev1.Secret{
		Type: corev1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret-test",
			Namespace: "ns",
		},
		Data: map[string][]byte{
			"some-key": []byte(data),
		},
	})
	if err != nil {
		t.Errorf("Unexpeced error: %v", err.Error())
	}

	secret, err := k8s.ReadSecret("ns", "secret-test")

	assert.Nil(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, "some-value", string(secret["some-key"]))
}

func TestReadSecretNotFound(t *testing.T) {
	client := fake.NewSimpleClientset()

	k8s := kubernetes.New(client, log.New())

	secret, err := k8s.ReadSecret("ns", "secret-test")
	assert.EqualError(t, err, fmt.Sprintf("[%s] secret '%s/%s' not found", errors.K8sSecretNotFoundErrorType, "ns", "secret-test"))
	assert.Empty(t, secret)
}
