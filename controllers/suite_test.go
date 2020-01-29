/*

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
	"errors"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	secretsmanagerv1alpha1 "github.com/tuenti/secrets-manager/api/v1alpha1"
	"k8s.io/client-go/rest"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var r *SecretDefinitionReconciler
var testEnv *envtest.Environment
var mgr ctrl.Manager
var scheme *runtime.Scheme

type fakeBackendSecret struct {
	Path    string
	Key     string
	Content string
}

type fakeBackend struct {
	fakeSecrets []fakeBackendSecret
}

func newFakeBackend(fakeSecrets []fakeBackendSecret) fakeBackend {
	return fakeBackend{
		fakeSecrets: fakeSecrets,
	}
}

func (f fakeBackend) ReadSecret(path string, key string) (string, error) {
	for _, fakeSecret := range f.fakeSecrets {
		if fakeSecret.Path == path && fakeSecret.Key == key {
			return fakeSecret.Content, nil
		}
	}
	return "", errors.New("Not found")

}

func getReconciler() *SecretDefinitionReconciler {
	return r
}

func getConfig() *rest.Config {
	return cfg
}

func getScheme() *runtime.Scheme {
	return scheme
}

func TestSecretDefinitionController(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	namespaces := [...]string{"notwatched", "watched", "watched1", "watched2"}
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
	}
	var err error

	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	scheme = runtime.NewScheme()
	corev1.AddToScheme(scheme)
	secretsmanagerv1alpha1.AddToScheme(scheme)

	err = secretsmanagerv1alpha1.AddToScheme(scheme)
	Expect(err).ToNot(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	mgr, err = ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0",
		LeaderElection:     false,
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(mgr).ToNot(BeNil())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).ToNot(HaveOccurred())

	for _, ns := range namespaces {
		nsSpec := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ""}}
		nsSpec.Name = ns
		k8sClient.Create(context.Background(), nsSpec)
	}

	r = &SecretDefinitionReconciler{
		Backend: newFakeBackend([]fakeBackendSecret{
			{"secret/data/pathtosecret1", "value", "bG9yZW0gaXBzdW0gZG9ybWEK"},
		}),
		Client:               k8sClient,
		APIReader:            k8sClient,
		Log:                  logf.Log.WithName("controllers-test").WithName("SecretDefinition"),
		Ctx:                  context.Background(),
		ReconciliationPeriod: 1 * time.Second,
	}
	err = r.SetupWithManager(mgr, "testing")
	//Expect(err).ToNot(HaveOccurred())*/

	Expect(err).ToNot(HaveOccurred())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
