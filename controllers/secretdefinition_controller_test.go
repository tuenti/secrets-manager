package controllers

import (
	"context"
	"encoding/base64"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smv1alpha1 "github.com/tuenti/secrets-manager/api/v1alpha1"
	"github.com/tuenti/secrets-manager/errors"

	"reflect"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	encodedValue = "bG9yZW0gaXBzdW0gZG9ybWEK"
	decodedValue = "lorem ipsum dorma"
)

var _ = Describe("SecretsManager", func() {
	var (
		cfg *rest.Config
		r   *SecretDefinitionReconciler
		sd  = &smv1alpha1.SecretDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "secretdef-test",
			},
			Spec: smv1alpha1.SecretDefinitionSpec{
				Name: "secret-test",
				Type: "Opaque",
				KeysMap: map[string]smv1alpha1.DataSource{
					"foo": smv1alpha1.DataSource{
						Path:     "secret/data/pathtosecret1",
						Key:      "value",
						Encoding: "base64",
					},
				},
			},
		}
		sd2 = &smv1alpha1.SecretDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "secretdef-test2",
			},
			Spec: smv1alpha1.SecretDefinitionSpec{
				Name: "secret-test2",
				Type: "Opaque",
				KeysMap: map[string]smv1alpha1.DataSource{
					"foo2": smv1alpha1.DataSource{
						Path:     "secret/data/pathtosecret1",
						Key:      "value",
						Encoding: "base64",
					},
				},
			},
		}
		sdNotWatched = &smv1alpha1.SecretDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "notwatched",
				Name:      "secretdef-notwatched",
			},
			Spec: smv1alpha1.SecretDefinitionSpec{
				Name: "secret-notwatched",
				Type: "Opaque",
				KeysMap: map[string]smv1alpha1.DataSource{
					"notwatched": smv1alpha1.DataSource{
						Path:     "secret/data/pathtosecret1",
						Key:      "value",
						Encoding: "base64",
					},
				},
			},
		}
		sdWatched = &smv1alpha1.SecretDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "watched",
				Name:      "secretdef-watched",
			},
			Spec: smv1alpha1.SecretDefinitionSpec{
				Name: "secret-watched",
				Type: "Opaque",
				KeysMap: map[string]smv1alpha1.DataSource{
					"watched": smv1alpha1.DataSource{
						Path:     "secret/data/pathtosecret1",
						Key:      "value",
						Encoding: "base64",
					},
				},
			},
		}
		sdMultiWatched1 = &smv1alpha1.SecretDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "watched1",
				Name:      "secretdef-multi1",
			},
			Spec: smv1alpha1.SecretDefinitionSpec{
				Name: "secret-multi1",
				Type: "Opaque",
				KeysMap: map[string]smv1alpha1.DataSource{
					"multival1": smv1alpha1.DataSource{
						Path:     "secret/data/pathtosecret1",
						Key:      "value",
						Encoding: "base64",
					},
				},
			},
		}
		sdMultiWatched2 = &smv1alpha1.SecretDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "watched2",
				Name:      "secretdef-multi2",
			},
			Spec: smv1alpha1.SecretDefinitionSpec{
				Name: "secret-multi2",
				Type: "Opaque",
				KeysMap: map[string]smv1alpha1.DataSource{
					"multival2": smv1alpha1.DataSource{
						Path:     "secret/data/pathtosecret1",
						Key:      "value",
						Encoding: "base64",
					},
				},
			},
		}
		sdBackendSecretNotFound = &smv1alpha1.SecretDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "secretdef-beckend-secret-not-found",
			},
			Spec: smv1alpha1.SecretDefinitionSpec{
				Name: "secret-backend-secret-not-found",
				Type: "Opaque",
				KeysMap: map[string]smv1alpha1.DataSource{
					"foo3": smv1alpha1.DataSource{
						Path:     "secret/data/notfound",
						Key:      "value",
						Encoding: "base64",
					},
				},
			},
		}
		sdWrongEncoding = &smv1alpha1.SecretDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "secretdef-wrong-encoding",
			},
			Spec: smv1alpha1.SecretDefinitionSpec{
				Name: "secret-wrong-encoding",
				Type: "Opaque",
				KeysMap: map[string]smv1alpha1.DataSource{
					"foo4": smv1alpha1.DataSource{
						Path:     "secret/data/pathtosecret1",
						Key:      "value",
						Encoding: "base65",
					},
				},
			},
		}
		sdWithLabels = &smv1alpha1.SecretDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "secretdef-labels",
				Labels: map[string]string{
					"test.example.com/name": "test",
					"name":                  "secret-labels",
				},
			},
			Spec: smv1alpha1.SecretDefinitionSpec{
				Name: "secret-labels",
				Type: "Opaque",
				KeysMap: map[string]smv1alpha1.DataSource{
					"fooLabel": smv1alpha1.DataSource{
						Path:     "secret/data/pathtosecret1",
						Key:      "value",
						Encoding: "base64",
					},
				},
			},
		}
		sdExcludedNs = &smv1alpha1.SecretDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "secretdef-excluded-ns",
			},
			Spec: smv1alpha1.SecretDefinitionSpec{
				Name: "secret-excluded-ns",
				Type: "Opaque",
				KeysMap: map[string]smv1alpha1.DataSource{
					"fooExcludedNs": smv1alpha1.DataSource{
						Path:     "secret/data/pathtosecret1",
						Key:      "value",
						Encoding: "base64",
					},
				},
			},
		}
	)

	BeforeEach(func() {
		r = getReconciler()
		cfg = getConfig()
	})

	AfterEach(func() {
	})

	Context("SecretDefinitionReconciler.Reconcile", func() {
		It("Create a secretdefinition and read the secret", func() {
			decodedBytes, _ := base64.StdEncoding.DecodeString(encodedValue)
			err := r.Create(context.Background(), sd)
			Expect(err).To(BeNil())
			res, err2 := r.Reconcile(reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: sd.Namespace,
					Name:      sd.Name,
				},
			})

			Expect(res).ToNot(BeNil())
			Expect(err2).To(BeNil())

			data, err3 := r.getCurrentState("default", "secret-test")
			Expect(err3).To(BeNil())
			Expect(data).To(Equal(map[string][]byte{"foo": decodedBytes}))
		})

		It("Delete a secretdefinition should delete a secret", func() {
			decodedBytes, _ := base64.StdEncoding.DecodeString(encodedValue)
			err := r.Create(context.Background(), sd2)
			Expect(err).To(BeNil())
			res, err2 := r.Reconcile(reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: sd2.Namespace,
					Name:      sd2.Name,
				},
			})

			Expect(res).ToNot(BeNil())
			Expect(err2).To(BeNil())
			Expect(sd2.ObjectMeta.Finalizers).To(BeEmpty())

			data, err3 := r.getCurrentState("default", "secret-test2")
			Expect(err3).To(BeNil())
			Expect(data).To(Equal(map[string][]byte{"foo2": decodedBytes}))

			err4 := r.Delete(context.Background(), sd2)
			Expect(err4).To(BeNil())
			res, err5 := r.Reconcile(reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: sd2.Namespace,
					Name:      sd2.Name,
				},
			})
			Expect(err5).To(BeNil())
			data2, err6 := r.getCurrentState("default", "secret-test2")
			Expect(err6).ToNot(BeNil())
			Expect(data2).To(BeEmpty())
		})
		It("Create a secretdefinition with a secret not deployed in the backend", func() {
			err := r.Create(context.Background(), sdBackendSecretNotFound)
			Expect(err).To(BeNil())
			res, err2 := r.Reconcile(reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: sdBackendSecretNotFound.Namespace,
					Name:      sdBackendSecretNotFound.Name,
				},
			})
			Expect(err2).ToNot(BeNil())
			Expect(res).To(Equal(reconcile.Result{}))
		})
		It("Create a secretdefinition with a wrong encoding", func() {
			expectedErr := &errors.EncodingNotImplementedError{}
			err := r.Create(context.Background(), sdWrongEncoding)
			Expect(err).To(BeNil())
			res, err2 := r.Reconcile(reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: sdWrongEncoding.Namespace,
					Name:      sdWrongEncoding.Name,
				},
			})
			Expect(reflect.TypeOf(err2)).To(Equal(reflect.TypeOf(expectedErr)))
			Expect(res).To(Equal(reconcile.Result{}))
		})
		It("Create a secretdefinition and read the labels and annotations", func() {
			err := r.Create(context.Background(), sdWithLabels)
			Expect(err).To(BeNil())
			res, err2 := r.Reconcile(reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: sdWithLabels.Namespace,
					Name:      sdWithLabels.Name,
				},
			})
			Expect(res).ToNot(BeNil())
			Expect(err2).To(BeNil())

			_, err3 := r.getCurrentState("default", "secret-labels")
			Expect(err3).To(BeNil())

			reader := r.APIReader
			secret := &corev1.Secret{}
			err4 := reader.Get(r.Ctx, client.ObjectKey{
				Namespace: sdWithLabels.Namespace,
				Name:      "secret-labels",
			}, secret)
			Expect(err4).To(BeNil())

			labels := secret.GetObjectMeta().GetLabels()
			Expect(labels).To(Equal(map[string]string{
				"app.kubernetes.io/managed-by": "secrets-manager",
				"name":                         "secret-labels",
				"test.example.com/name":        "test"}))

			annotations := secret.GetObjectMeta().GetAnnotations()
			_, ok := annotations["secrets-manager.tuenti.io/lastUpdateTime"]
			Expect(ok).To(BeTrue())
		})
		It("Create a secretdefinition in a non-watched namespace", func() {
			r2 := getReconciler()
			r2.WatchNamespaces = map[string]bool{"watch": true}
			err := r.Create(context.Background(), sd3)
			res, err2 := r.Reconcile(reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: sd3.Namespace,
					Name:      sd3.Name,
				},
			})
			Expect(err).To(BeNil())
			Expect(err2).To(BeNil())
			Expect(res).To(Equal(reconcile.Result{}))
		})
		It("Create a secretdefinition in a watched namespace", func() {
			decodedBytes, _ := base64.StdEncoding.DecodeString(encodedValue)
			r2 := getReconciler()
			r2.WatchNamespaces = map[string]bool{sd4.Namespace: true}
			err := r.Create(context.Background(), sd4)
			res, err2 := r.Reconcile(reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: sd4.Namespace,
					Name:      sd4.Name,
				},
			})
			Expect(err).To(BeNil())
			Expect(err2).To(BeNil())
			Expect(res).ToNot(BeNil())

			data, err3 := r.getCurrentState("default", "secret-test4")
			Expect(err3).To(BeNil())
			Expect(data).To(Equal(map[string][]byte{"foo4": decodedBytes}))
		})
		It("Create a secretdefinition in a excluded namespace", func() {
			r2 := getReconciler()
			r2.ExcludeNamespaces = map[string]bool{sdExcludedNs.Namespace: true}
			err := r.Create(context.Background(), sdExcludedNs)
			res, err2 := r.Reconcile(reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: sdExcludedNs.Namespace,
					Name:      sdExcludedNs.Name,
				},
			})
			Expect(err).To(BeNil())
			Expect(err2).To(BeNil())
			Expect(res).To(Equal(reconcile.Result{}))
		})
	})
	Context("SecretDefinitionReconciler.upsertSecret", func() {
		It("Upsert a secret twice should not raise an error", func() {
			decodedBytes, _ := base64.StdEncoding.DecodeString(encodedValue)
			err := r.upsertSecret(sd, map[string][]byte{"foo": decodedBytes})
			Expect(err).To(BeNil())
			err2 := r.upsertSecret(sd, map[string][]byte{"foo": decodedBytes})
			Expect(err2).To(BeNil())
		})
	})
	Context("Manager.MultiNamespacedCache", func() {

		It("Creates secret in watched namespace", func(done Done) {
			scheme := getScheme()

			decodedBytes, _ := base64.StdEncoding.DecodeString(encodedValue)
			mgr, err := ctrl.NewManager(cfg, ctrl.Options{
				Scheme:             scheme,
				MetricsBindAddress: "0",
				LeaderElection:     false,
				NewCache:           cache.MultiNamespacedCacheBuilder([]string{"watched"}),
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(mgr).ToNot(BeNil())
			r2 := getReconciler()
			r2.SetupWithManager(mgr, "test1")

			c1 := make(chan struct{})
			go func() {
				defer GinkgoRecover()
				Expect(mgr.Start(c1)).NotTo(HaveOccurred())
				close(done)
			}()

			r2.Create(context.Background(), sdWatched)
			// Sleep for 4 * the reconcile interval set on the controller (just to be safe)
			time.Sleep(4 * time.Second)
			data, err := r2.getCurrentState("watched", sdWatched.Spec.Name)
			Expect(err).To(BeNil())
			Expect(data).To(Equal(map[string][]byte{"watched": decodedBytes}))
			close(c1)

		}, 10)

		It("Doesn't create secret in unwatched namespace", func(done Done) {
			scheme := getScheme()

			mgr, err := ctrl.NewManager(cfg, ctrl.Options{
				Scheme:             scheme,
				MetricsBindAddress: "0",
				LeaderElection:     false,
				NewCache:           cache.MultiNamespacedCacheBuilder([]string{"watched"}),
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(mgr).ToNot(BeNil())
			r2 := getReconciler()
			r2.SetupWithManager(mgr, "test2")

			c1 := make(chan struct{})
			go func() {
				defer GinkgoRecover()
				Expect(mgr.Start(c1)).NotTo(HaveOccurred())
				close(done)
			}()

			r2.Create(context.Background(), sdNotWatched)
			// Sleep for 4 * the reconcile interval set on the controller (just to be safe)
			time.Sleep(4 * time.Second)
			data, err := r2.getCurrentState("notwatched", sdNotWatched.Spec.Name)
			Expect(err.Error()).To(Equal("secrets \"secret-notwatched\" not found"))
			Expect(data).To(BeEmpty())
			close(c1)

		}, 10)

		It("Creates secrets in multiple watched namespaces", func(done Done) {
			scheme := getScheme()

			decodedBytes, _ := base64.StdEncoding.DecodeString(encodedValue)
			mgr, err := ctrl.NewManager(cfg, ctrl.Options{
				Scheme:             scheme,
				MetricsBindAddress: "0",
				LeaderElection:     false,
				NewCache:           cache.MultiNamespacedCacheBuilder([]string{"watched1", "watched2"}),
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(mgr).ToNot(BeNil())
			r2 := getReconciler()
			r2.SetupWithManager(mgr, "test3")

			c1 := make(chan struct{})
			go func() {
				defer GinkgoRecover()
				Expect(mgr.Start(c1)).NotTo(HaveOccurred())
				close(done)
			}()

			r2.Create(context.Background(), sdMultiWatched1)
			r2.Create(context.Background(), sdMultiWatched2)
			// Sleep for 4 * the reconcile interval set on the controller (just to be safe)
			time.Sleep(4 * time.Second)
			data, err2 := r2.getCurrentState("watched1", sdMultiWatched1.Spec.Name)
			Expect(err2).To(BeNil())
			Expect(data).To(Equal(map[string][]byte{"multival1": decodedBytes}))

			data2, err3 := r2.getCurrentState("watched2", sdMultiWatched2.Spec.Name)
			Expect(err3).To(BeNil())
			Expect(data2).To(Equal(map[string][]byte{"multival2": decodedBytes}))

			close(c1)

		}, 10)
	})
})
