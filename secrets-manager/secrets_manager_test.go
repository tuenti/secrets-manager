package secretsmanager

import (
	"context"
	"errors"
	"fmt"

	"testing"

	gomock "github.com/golang/mock/gomock"

	"github.com/stretchr/testify/assert"
	e "github.com/tuenti/secrets-manager/errors"
	"github.com/tuenti/secrets-manager/kubernetes"
	"github.com/tuenti/secrets-manager/mocks"
	"k8s.io/client-go/kubernetes/fake"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

type fakeBackendSecret struct {
	Path    string
	Key     string
	Content string
}

type fakeBackend struct {
	fakeSecrets []fakeBackendSecret
}

func (f fakeBackend) ReadSecret(path string, key string) (string, error) {
	for _, fakeSecret := range f.fakeSecrets {
		if fakeSecret.Path == path && fakeSecret.Key == key {
			return fakeSecret.Content, nil
		}
	}
	return "", errors.New("Not found")
}

func newFakeBackend(fakeSecrets []fakeBackendSecret) fakeBackend {
	return fakeBackend{
		fakeSecrets: fakeSecrets,
	}
}

func TestNew(t *testing.T) {
	ctx := context.Background()
	fakeBackend := newFakeBackend([]fakeBackendSecret{})
	logger := log.New()
	k8s := kubernetes.New(fake.NewSimpleClientset(), logger)
	cfg := Config{ConfigMap: "cm"}
	secretManager, err := New(ctx, cfg, k8s, fakeBackend, logger)

	assert.Nil(t, err)
	assert.NotNil(t, secretManager)
	assert.Equal(t, "default", secretManager.configMapNamespace)
	assert.Equal(t, "cm", secretManager.configMapName)
}

func TestNewSplitConfigMapName(t *testing.T) {
	ctx := context.Background()
	fakeBackend := newFakeBackend([]fakeBackendSecret{})
	logger := log.New()
	k8s := kubernetes.New(fake.NewSimpleClientset(), logger)
	cfg := Config{ConfigMap: "ns/cm"}
	secretManager, err := New(ctx, cfg, k8s, fakeBackend, logger)

	assert.Nil(t, err)
	assert.NotNil(t, secretManager)
	assert.Equal(t, "ns", secretManager.configMapNamespace)
	assert.Equal(t, "cm", secretManager.configMapName)
}

func TestNewConfigMapNameMultipleSlashes(t *testing.T) {
	ctx := context.Background()
	fakeBackend := newFakeBackend([]fakeBackendSecret{})
	logger := log.New()
	k8s := kubernetes.New(fake.NewSimpleClientset(), logger)
	cfg := Config{ConfigMap: "ns/cm/foo"}
	secretManager, err := New(ctx, cfg, k8s, fakeBackend, logger)

	assert.NotNil(t, err)
	assert.Nil(t, secretManager)
}

func TestGetDesiredState(t *testing.T) {
	ctx := context.Background()
	fakeBackend := newFakeBackend([]fakeBackendSecret{
		{"some/path", "key-in-vault", "fake-content"},
	})
	logger := log.New()
	k8s := kubernetes.New(fake.NewSimpleClientset(), logger)
	cfg := Config{ConfigMap: "cm"}
	secretManager, _ := New(ctx, cfg, k8s, fakeBackend, logger)

	data, err := secretManager.getDesiredState(SecretDefinition{
		Data: map[string]Datasource{
			"key1": {
				Path: "some/path",
				Key:  "key-in-vault",
			},
		},
	})

	assert.Nil(t, err)
	assert.Len(t, data, 1)
}

func TestGetDesiredStateBadB64Content(t *testing.T) {
	ctx := context.Background()
	fakeBackend := newFakeBackend([]fakeBackendSecret{
		{"some/path", "key-in-vault", "this is not base64!!"},
	})
	logger := log.New()
	k8s := kubernetes.New(fake.NewSimpleClientset(), logger)
	cfg := Config{ConfigMap: "cm"}
	secretManager, _ := New(ctx, cfg, k8s, fakeBackend, logger)

	data, err := secretManager.getDesiredState(SecretDefinition{
		Data: map[string]Datasource{
			"key1": {
				Encoding: "base64",
				Path:     "some/path",
				Key:      "key-in-vault",
			},
		},
	})

	assert.Nil(t, data)
	assert.NotNil(t, err)
}

func TestGetDesiredStateEncodingNotImplemented(t *testing.T) {
	ctx := context.Background()
	fakeBackend := newFakeBackend([]fakeBackendSecret{
		{"some/path", "key-in-vault", "fake-content"},
	})
	logger := log.New()
	k8s := kubernetes.New(fake.NewSimpleClientset(), logger)
	cfg := Config{ConfigMap: "cm"}
	secretManager, _ := New(ctx, cfg, k8s, fakeBackend, logger)

	data, err := secretManager.getDesiredState(SecretDefinition{
		Data: map[string]Datasource{
			"key1": {
				Encoding: "base65",
				Path:     "some/path",
				Key:      "key-in-vault",
			},
		},
	})
	assert.Nil(t, data)
	assert.NotNil(t, err)
	assert.EqualError(t, err, fmt.Sprintf("[%s] encoding %s not supported", e.EncodingNotImplementedErrorType, "base65"))
}

func TestGetDesiredStateBackendError(t *testing.T) {
	ctx := context.Background()
	fakeBackend := newFakeBackend([]fakeBackendSecret{})
	logger := log.New()
	k8s := kubernetes.New(fake.NewSimpleClientset(), logger)
	cfg := Config{ConfigMap: "cm"}
	secretManager, _ := New(ctx, cfg, k8s, fakeBackend, logger)

	_, err := secretManager.getDesiredState(SecretDefinition{
		Data: map[string]Datasource{
			"key1": {
				Path: "some/path",
				Key:  "key-in-vault",
			},
		},
	})

	assert.NotNil(t, err)
}

func TestGetCurrentState(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	k8s := mocks.NewMockKubernetesClient(mockCtrl)
	fakeSecretData := map[string][]byte{
		"value1": []byte("Fake Value"),
	}
	k8s.EXPECT().ReadSecret("ns", "secret-name").AnyTimes().Return(fakeSecretData, nil)

	ctx := context.Background()
	fakeBackend := newFakeBackend([]fakeBackendSecret{})
	logger := log.New()
	cfg := Config{ConfigMap: "cm"}
	secretManager, _ := New(ctx, cfg, k8s, fakeBackend, logger)

	data, err := secretManager.getCurrentState("ns", "secret-name")

	assert.Nil(t, err)
	assert.Equal(t, fakeSecretData, data)
}

func TestGetCurrentStateError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	k8s := mocks.NewMockKubernetesClient(mockCtrl)
	k8s.EXPECT().ReadSecret("ns", "secret-name").AnyTimes().Return(nil, errors.New("some-error"))

	ctx := context.Background()
	fakeBackend := newFakeBackend([]fakeBackendSecret{})
	logger := log.New()
	cfg := Config{ConfigMap: "cm"}
	secretManager, _ := New(ctx, cfg, k8s, fakeBackend, logger)

	_, err := secretManager.getCurrentState("ns", "secret-name")

	assert.NotNil(t, err)
}

func TestUpsertSecret(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	k8s := mocks.NewMockKubernetesClient(mockCtrl)

	expectedSecret := &kubernetes.Secret{
		Name:      "secret-name",
		Namespace: "ns",
		Data: map[string][]byte{
			"value1": []byte("fake-data"),
		},
	}

	k8s.EXPECT().UpsertSecret(EqSecret(expectedSecret)).Times(1).Return(nil)

	ctx := context.Background()
	fakeBackend := newFakeBackend([]fakeBackendSecret{})
	logger := log.New()
	cfg := Config{ConfigMap: "cm"}
	secretManager, _ := New(ctx, cfg, k8s, fakeBackend, logger)

	err := secretManager.upsertSecret(
		"Opaque",
		"ns",
		"secret-name",
		map[string][]byte{
			"value1": []byte("fake-data"),
		})

	assert.Nil(t, err)
}

func TestUpsertSecretError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	k8s := mocks.NewMockKubernetesClient(mockCtrl)

	expectedSecret := &kubernetes.Secret{
		Name:      "secret-name",
		Namespace: "ns",
		Data: map[string][]byte{
			"value1": []byte("fake-data"),
		},
	}

	k8s.EXPECT().UpsertSecret(EqSecret(expectedSecret)).Times(1).Return(errors.New("some-error"))

	ctx := context.Background()
	fakeBackend := newFakeBackend([]fakeBackendSecret{})
	logger := log.New()
	cfg := Config{ConfigMap: "cm"}
	secretManager, _ := New(ctx, cfg, k8s, fakeBackend, logger)

	err := secretManager.upsertSecret(
		"Opaque",
		"ns",
		"secret-name",
		map[string][]byte{
			"value1": []byte("fake-data"),
		})

	assert.NotNil(t, err)
}

func TestSyncState(t *testing.T) {
	secretSyncErrorsCount.Reset()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	k8s := mocks.NewMockKubernetesClient(mockCtrl)

	fakeCurrentSecretData := map[string][]byte{
		"value1": []byte("fake-current-data"),
	}

	expectedSecret := &kubernetes.Secret{
		Name:      "secret-name",
		Namespace: "ns",
		Data: map[string][]byte{
			"value1": []byte("fake-content"),
		},
	}

	k8s.EXPECT().ReadSecret("ns", "secret-name").AnyTimes().Return(fakeCurrentSecretData, nil)
	k8s.EXPECT().UpsertSecret(EqSecret(expectedSecret)).Times(1).Return(nil)

	ctx := context.Background()
	fakeBackend := newFakeBackend([]fakeBackendSecret{
		{"some/path", "key-in-vault", "fake-content"},
	})
	logger := log.New()
	cfg := Config{ConfigMap: "cm"}
	secretManager, _ := New(ctx, cfg, k8s, fakeBackend, logger)

	err := secretManager.syncState(SecretDefinition{
		Name:       "secret-name",
		Namespaces: []string{"ns"},
		Type:       "Opaque",
		Data: map[string]Datasource{
			"value1": {
				Path: "some/path",
				Key:  "key-in-vault",
			},
		},
	})

	assert.Nil(t, err)
	// Test Prometheus metric
	metricSecretSyncErrorsCount, _ := secretSyncErrorsCount.GetMetricWithLabelValues("secret-name", "ns")
	assert.Equal(t, 0.0, testutil.ToFloat64(metricSecretSyncErrorsCount))
}

func TestSyncStateErrorGetDesired(t *testing.T) {
	secretSyncErrorsCount.Reset()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	k8s := mocks.NewMockKubernetesClient(mockCtrl)

	ctx := context.Background()
	fakeBackend := newFakeBackend([]fakeBackendSecret{})
	logger := log.New()
	cfg := Config{ConfigMap: "cm"}
	secretManager, _ := New(ctx, cfg, k8s, fakeBackend, logger)

	err := secretManager.syncState(SecretDefinition{
		Name:       "secret-name",
		Namespaces: []string{"ns"},
		Type:       "Opaque",
		Data: map[string]Datasource{
			"value1": {
				Path: "some/path",
				Key:  "key-in-vault",
			},
		},
	})

	assert.NotNil(t, err)
	// Test Prometheus metric
	metricSecretSyncErrorsCount, _ := secretSyncErrorsCount.GetMetricWithLabelValues("secret-name", "ns")
	assert.Equal(t, 1.0, testutil.ToFloat64(metricSecretSyncErrorsCount))
}

func TestSyncStateErrorGetCurrentInOneSecret(t *testing.T) {
	secretSyncErrorsCount.Reset()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	k8s := mocks.NewMockKubernetesClient(mockCtrl)

	fakeCurrentSecretData := map[string][]byte{
		"value1": []byte("fake-current-data"),
	}

	expectedSecret1 := &kubernetes.Secret{
		Name:      "secret-name",
		Namespace: "ns1",
		Data: map[string][]byte{
			"value1": []byte("fake-content"),
		},
	}

	expectedSecret3 := &kubernetes.Secret{
		Name:      "secret-name",
		Namespace: "ns3",
		Data: map[string][]byte{
			"value1": []byte("fake-content"),
		},
	}

	k8s.EXPECT().ReadSecret("ns1", "secret-name").AnyTimes().Return(fakeCurrentSecretData, nil)
	k8s.EXPECT().ReadSecret("ns2", "secret-name").AnyTimes().Return(nil, errors.New("some error"))
	k8s.EXPECT().ReadSecret("ns3", "secret-name").AnyTimes().Return(fakeCurrentSecretData, nil)
	k8s.EXPECT().UpsertSecret(EqSecret(expectedSecret1)).Times(1).Return(nil)
	k8s.EXPECT().UpsertSecret(EqSecret(expectedSecret3)).Times(1).Return(nil)

	ctx := context.Background()
	fakeBackend := newFakeBackend([]fakeBackendSecret{
		{"some/path", "key-in-vault", "fake-content"},
	})
	logger := log.New()
	cfg := Config{ConfigMap: "cm"}
	secretManager, _ := New(ctx, cfg, k8s, fakeBackend, logger)

	err := secretManager.syncState(SecretDefinition{
		Name:       "secret-name",
		Namespaces: []string{"ns1", "ns2", "ns3"},
		Type:       "Opaque",
		Data: map[string]Datasource{
			"value1": {
				Path: "some/path",
				Key:  "key-in-vault",
			},
		},
	})

	assert.Nil(t, err)
	// Test Prometheus metric
	metricSecretSyncErrorsCount1, _ := secretSyncErrorsCount.GetMetricWithLabelValues("secret-name", "ns1")
	assert.Equal(t, 0.0, testutil.ToFloat64(metricSecretSyncErrorsCount1))
	// Test Prometheus metric
	metricSecretSyncErrorsCount2, _ := secretSyncErrorsCount.GetMetricWithLabelValues("secret-name", "ns2")
	assert.Equal(t, 1.0, testutil.ToFloat64(metricSecretSyncErrorsCount2))
	// Test Prometheus metric
	metricSecretSyncErrorsCount3, _ := secretSyncErrorsCount.GetMetricWithLabelValues("secret-name", "ns3")
	assert.Equal(t, 0.0, testutil.ToFloat64(metricSecretSyncErrorsCount3))
}

func TestSyncStateErrorUpsertSecret(t *testing.T) {
	secretSyncErrorsCount.Reset()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	k8s := mocks.NewMockKubernetesClient(mockCtrl)

	fakeCurrentSecretData := map[string][]byte{
		"value1": []byte("fake-current-data"),
	}

	expectedSecret1 := &kubernetes.Secret{
		Name:      "secret-name",
		Namespace: "ns1",
		Data: map[string][]byte{
			"value1": []byte("fake-content"),
		},
	}


	k8s.EXPECT().ReadSecret("ns1", "secret-name").AnyTimes().Return(fakeCurrentSecretData, nil)
	k8s.EXPECT().UpsertSecret(EqSecret(expectedSecret1)).Times(1).Return(errors.New("some error"))

	ctx := context.Background()
	fakeBackend := newFakeBackend([]fakeBackendSecret{
		{"some/path", "key-in-vault", "fake-content"},
	})
	logger := log.New()
	cfg := Config{ConfigMap: "cm"}
	secretManager, _ := New(ctx, cfg, k8s, fakeBackend, logger)

	err := secretManager.syncState(SecretDefinition{
		Name:       "secret-name",
		Namespaces: []string{"ns1"},
		Type:       "Opaque",
		Data: map[string]Datasource{
			"value1": {
				Path: "some/path",
				Key:  "key-in-vault",
			},
		},
	})

	assert.Nil(t, err)
	// Test Prometheus metric
	metricSecretSyncErrorsCount, _ := secretSyncErrorsCount.GetMetricWithLabelValues("secret-name", "ns")
	assert.Equal(t, 0.0, testutil.ToFloat64(metricSecretSyncErrorsCount))
}

func TestLoadConfig(t *testing.T) {
	configText := `
- name: supersecret1
  type: kubernetes.io/tls
  namespaces:
  - default
  data:
    tls.crt:
      path: secret/data/pathtosecret1
      key: value
    tls.key:
      path: secret/data/pathtosecret2
      key: value

- name: supersecret2
  type: Opaque
  namespaces:
  - default
  data:
    value1:
      path: secret/data/pathtosecret1
      key: value
    value2:
      path: secret/data/pathtosecret1
      key: value`

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ctx := context.Background()
	fakeBackend := newFakeBackend([]fakeBackendSecret{})
	logger := log.New()
	k8s := mocks.NewMockKubernetesClient(mockCtrl)
	cfg := Config{ConfigMap: "cm"}

	k8s.EXPECT().ReadConfigMap("cm", "default", "secretDefinitions").AnyTimes().Return(configText, nil)

	secretManager, _ := New(ctx, cfg, k8s, fakeBackend, logger)

	err := secretManager.loadSecretDefinitions()

	assert.Nil(t, err)
	assert.Len(t, secretManager.secretDefinitions, 2)
}

func TestLoadConfigConfigMapNotFound(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ctx := context.Background()
	fakeBackend := newFakeBackend([]fakeBackendSecret{})
	logger := log.New()
	k8s := mocks.NewMockKubernetesClient(mockCtrl)
	cfg := Config{ConfigMap: "cm"}

	k8s.EXPECT().ReadConfigMap("cm", "default", "secretDefinitions").AnyTimes().Return("", errors.New("not found"))

	secretManager, _ := New(ctx, cfg, k8s, fakeBackend, logger)

	err := secretManager.loadSecretDefinitions()

	assert.NotNil(t, err)
}

func TestLoadConfigConfigMapBadYaml(t *testing.T) {
	configText := `This is Bad Yaml`
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ctx := context.Background()
	fakeBackend := newFakeBackend([]fakeBackendSecret{})
	logger := log.New()
	k8s := mocks.NewMockKubernetesClient(mockCtrl)
	cfg := Config{ConfigMap: "cm"}

	k8s.EXPECT().ReadConfigMap("cm", "default", "secretDefinitions").AnyTimes().Return(configText, nil)

	secretManager, _ := New(ctx, cfg, k8s, fakeBackend, logger)

	err := secretManager.loadSecretDefinitions()

	assert.NotNil(t, err)
}
