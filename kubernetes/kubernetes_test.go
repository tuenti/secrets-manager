package kubernetes

import (
	"errors"
	"fmt"
	"os"

	secretsManagerErrors "github.com/tuenti/secrets-manager/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	clientgotesting "k8s.io/client-go/testing"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestUpsertSecretDoesNotExist(t *testing.T) {
	secretUpdateErrorsTotal.Reset()
	client := fake.NewSimpleClientset()

	k8s := New(client, log.New())

	k8sSecret := NewFakeSecret("ns", "secret-test")

	k8s.UpsertSecret(k8sSecret)

	secret, _ := client.CoreV1().Secrets("ns").Get("secret-test", metav1.GetOptions{})

	assert.NotNil(t, secret)
	metricSecretUpdateErrorsTotal, _ := secretUpdateErrorsTotal.GetMetricWithLabelValues("secret-test", "ns")
	assert.Equal(t, 0.0, testutil.ToFloat64(metricSecretUpdateErrorsTotal))
}

func TestUpsertSecretAlreadyExists(t *testing.T) {
	secretUpdateErrorsTotal.Reset()
	// Create the fake client.
	client := fake.NewSimpleClientset()

	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	logger.Out = os.Stderr

	k8s := New(client, logger)

	k8sSecret := NewFakeSecret("ns", "secret-test")

	// Upsert twice, second must be an update
	k8s.UpsertSecret(k8sSecret)
	k8s.UpsertSecret(k8sSecret)

	actions := client.Actions()
	lastAction := actions[len(actions)-1]
	assert.Implements(t, (*clientgotesting.UpdateAction)(nil), lastAction, "Last action must be UpdateAction")
	metricSecretUpdateErrorsTotal, _ := secretUpdateErrorsTotal.GetMetricWithLabelValues("secret-test", "ns")
	assert.Equal(t, 0.0, testutil.ToFloat64(metricSecretUpdateErrorsTotal))
}

func TestReadConfigMap(t *testing.T) {

	// Create the fake client.
	client := fake.NewSimpleClientset()

	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	logger.Out = os.Stderr

	k8s := New(client, logger)

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
	secretReadErrorsTotal.Reset()
	client := fake.NewSimpleClientset()

	k8s := New(client, log.New())

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
	metricSecretReadErrorsTotal, _ := secretUpdateErrorsTotal.GetMetricWithLabelValues("secret-test", "ns")
	assert.Equal(t, 0.0, testutil.ToFloat64(metricSecretReadErrorsTotal))
}

func TestReadSecretNotFound(t *testing.T) {
	secretReadErrorsTotal.Reset()
	client := fake.NewSimpleClientset()

	k8s := New(client, log.New())

	secret, err := k8s.ReadSecret("ns", "secret-test")
	assert.EqualError(t, err, fmt.Sprintf("[%s] secret '%s/%s' not found", secretsManagerErrors.K8sSecretNotFoundErrorType, "ns", "secret-test"))
	assert.Empty(t, secret)
	metricSecretReadErrorsTotal, _ := secretUpdateErrorsTotal.GetMetricWithLabelValues("secret-test", "ns")
	assert.Equal(t, 0.0, testutil.ToFloat64(metricSecretReadErrorsTotal))
}

func TestReadSecretError(t *testing.T) {
	secretReadErrorsTotal.Reset()

	client := fake.NewSimpleClientset()
	client.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("*", "*", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.New("This is an This is an unexpected K8s error")
	})

	k8s := New(client, log.New())
	secret, err := k8s.ReadSecret("ns", "secret-test")
	assert.Empty(t, secret)
	assert.NotNil(t, err)
	metricSecretReadErrorsTotal, _ := secretReadErrorsTotal.GetMetricWithLabelValues("secret-test", "ns")
	assert.Equal(t, 1.0, testutil.ToFloat64(metricSecretReadErrorsTotal))
}
