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

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	encodedValue = "bG9yZW0gaXBzdW0gZG9ybWEK"
	decodedValue = "lorem ipsum dorma"
)

var _ = Describe("SecretsManager", func() {
	var (
		//cfg *rest.Config
		r *SecretDefinitionReconciler

		decodedBytes, _ = base64.StdEncoding.DecodeString(encodedValue)
		anyData         = map[string][]byte{"foo": decodedBytes}

		sd = &smv1alpha1.SecretDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "secret-test",
				Labels: map[string]string{
					"test.example.com/name": "test",
					"name":                  "secret_labels",
				},
				Annotations: map[string]string{
					"ann1": "another_value",
					"ann2": "just_a_value",
				},
			},
			Spec: smv1alpha1.SecretDefinitionSpec{
				Name: "secret-test",
				Type: "Opaque",
				KeysMap: map[string]smv1alpha1.DataSource{
					"foo": {
						Path:     "secret/data/pathtosecret1",
						Key:      "value",
						Encoding: "base64",
					},
				},
			},
		}
		sdWithSkipAnnotations = &smv1alpha1.SecretDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "secret-test",
				Labels: map[string]string{
					"test.example.com/name": "test",
					"name":                  "secret_labels",
				},
				Annotations: map[string]string{
					"ann1":                             "another_value",
					"ann2":                             "just_a_value",
					corev1.LastAppliedConfigAnnotation: "to_be_skipped",
				},
			},
			Spec: smv1alpha1.SecretDefinitionSpec{
				Name: "secret-test",
				Type: "Opaque",
				KeysMap: map[string]smv1alpha1.DataSource{
					"foo": {
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
				Name:      "secret-test2",
			},
			Spec: smv1alpha1.SecretDefinitionSpec{
				Name: "secret-test2",
				Type: "Opaque",
				KeysMap: map[string]smv1alpha1.DataSource{
					"foo2": {
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
				Name:      "secret-notwatched",
			},
			Spec: smv1alpha1.SecretDefinitionSpec{
				Name: "secret-notwatched",
				Type: "Opaque",
				KeysMap: map[string]smv1alpha1.DataSource{
					"notwatched": {
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
				Name:      "secret-watched",
			},
			Spec: smv1alpha1.SecretDefinitionSpec{
				Name: "secret-watched",
				Type: "Opaque",
				KeysMap: map[string]smv1alpha1.DataSource{
					"watched": {
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
				Name:      "secret-multi1",
			},
			Spec: smv1alpha1.SecretDefinitionSpec{
				Name: "secret-multi1",
				Type: "Opaque",
				KeysMap: map[string]smv1alpha1.DataSource{
					"multival1": {
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
				Name:      "secret-multi2",
			},
			Spec: smv1alpha1.SecretDefinitionSpec{
				Name: "secret-multi2",
				Type: "Opaque",
				KeysMap: map[string]smv1alpha1.DataSource{
					"multival2": {
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
				Name:      "secret-beckend-secret-not-found",
			},
			Spec: smv1alpha1.SecretDefinitionSpec{
				Name: "secret-backend-secret-not-found",
				Type: "Opaque",
				KeysMap: map[string]smv1alpha1.DataSource{
					"foo3": {
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
				Name:      "secret-wrong-encoding",
			},
			Spec: smv1alpha1.SecretDefinitionSpec{
				Name: "secret-wrong-encoding",
				Type: "Opaque",
				KeysMap: map[string]smv1alpha1.DataSource{
					"foo4": {
						Path:     "secret/data/pathtosecret1",
						Key:      "value",
						Encoding: "base65",
					},
				},
			},
		}
		sdExcludedNs = &smv1alpha1.SecretDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "secret-excluded-ns",
			},
			Spec: smv1alpha1.SecretDefinitionSpec{
				Name: "secret-excluded-ns",
				Type: "Opaque",
				KeysMap: map[string]smv1alpha1.DataSource{
					"fooExcludedNs": {
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
			// setup:
			secretdefinition := sd
			ctx := context.Background()

			// when:
			err := r.Create(ctx, secretdefinition)
			res, err2 := r.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: secretdefinition.Namespace,
					Name:      secretdefinition.Name,
				},
			})
			data, err3 := r.getCurrentState(ctx, "default", secretdefinition.ObjectMeta.Name)

			// then:
			Expect(err).To(BeNil())

			Expect(res).ToNot(BeNil())
			Expect(err2).To(BeNil())

			Expect(err3).To(BeNil())
			Expect(data).To(Equal(anyData))
		})

		It("Delete a secretdefinition should delete a secret", func() {
			// setup:
			secretdefinition := sd2
			ctx := context.Background()

			// when:
			err := r.Create(ctx, secretdefinition)
			res, err2 := r.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: secretdefinition.Namespace,
					Name:      secretdefinition.Name,
				},
			})

			// then:
			Expect(err).To(BeNil())
			Expect(res).ToNot(BeNil())
			Expect(err2).To(BeNil())
			Expect(secretdefinition.ObjectMeta.Finalizers).To(BeEmpty())

			// when:
			data, err3 := r.getCurrentState(ctx, "default", secretdefinition.ObjectMeta.Name)

			// then:
			Expect(err3).To(BeNil())
			Expect(data).To(Equal(map[string][]byte{"foo2": decodedBytes}))

			// when:
			err4 := r.Delete(ctx, secretdefinition)
			res, err5 := r.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: secretdefinition.Namespace,
					Name:      secretdefinition.Name,
				},
			})
			data2, err6 := r.getCurrentState(ctx, "default", secretdefinition.ObjectMeta.Name)

			// then:
			Expect(err4).To(BeNil())
			Expect(err5).To(BeNil())
			Expect(err6).ToNot(BeNil())
			Expect(data2).To(BeEmpty())

			// when:
			res, err10 := r.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: secretdefinition.Namespace,
					Name:      secretdefinition.Name,
				},
			})
			Expect(err10).To(BeNil())
		})
		It("Create a secretdefinition with a secret not deployed in the backend", func() {
			ctx := context.Background()
			err := r.Create(ctx, sdBackendSecretNotFound)
			Expect(err).To(BeNil())
			res, err2 := r.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: sdBackendSecretNotFound.Namespace,
					Name:      sdBackendSecretNotFound.Name,
				},
			})
			Expect(err2).ToNot(BeNil())
			Expect(res).To(Equal(reconcile.Result{}))
		})
		It("Create a secretdefinition with a wrong encoding", func() {
			ctx := context.Background()
			expectedErr := &errors.EncodingNotImplementedError{}
			err := r.Create(ctx, sdWrongEncoding)
			Expect(err).To(BeNil())
			res, err2 := r.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: sdWrongEncoding.Namespace,
					Name:      sdWrongEncoding.Name,
				},
			})
			Expect(reflect.TypeOf(err2)).To(Equal(reflect.TypeOf(expectedErr)))
			Expect(res).To(Equal(reconcile.Result{}))
		})
		It("Create a secretdefinition in a excluded namespace", func() {
			// setup:
			secretdefinition := sdExcludedNs
			r2 := getReconciler()
			r2.ExcludeNamespaces = map[string]bool{secretdefinition.Namespace: true}
			ctx := context.Background()

			// when:
			err := r.Create(ctx, secretdefinition)
			res, err2 := r.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: secretdefinition.Namespace,
					Name:      secretdefinition.Name,
				},
			})

			// then:
			Expect(err).To(BeNil())

			Expect(err2).To(BeNil())
			Expect(res).To(Equal(reconcile.Result{}))
		})
	})
	Context("SecretDefinitionReconciler.upsertSecret", func() {

		It("Upsert a secret twice should not raise an error", func() {
			// setup:
			secretdefinition := sd
			ctx := context.Background()

			// when:
			err := r.upsertSecret(ctx, secretdefinition, anyData)
			err2 := r.upsertSecret(ctx, secretdefinition, anyData)

			// then:
			Expect(err).To(BeNil())
			Expect(err2).To(BeNil())
		})
		It("Upsert a secret", func() {
			// setup:
			secretdefinition := sd
			ctx := context.Background()

			// when:
			err := r.upsertSecret(ctx, secretdefinition, anyData)

			// then:
			Expect(err).To(BeNil())
		})
	})
	Context("SecretDefinitionReconciler.getObjectMetaFromSecretDefinition", func() {

		It("getObjectMetaFromSecretDefinition should return an ObjectMeta", func() {
			// when:
			objectMeta := getObjectMetaFromSecretDefinition(sd)

			// then:
			Expect(sd.Namespace).To(Equal(objectMeta.Namespace))
			Expect(sd.Name).To(Equal(objectMeta.Name))

			for k := range sd.Labels {
				Expect(objectMeta.Labels).Should(HaveKey(k))
			}
			for k := range sd.Annotations {
				Expect(objectMeta.Annotations).Should(HaveKey(k))
			}
		})
		It("getObjectMetaFromSecretDefinition should add custom labels and annotations to the objectMeta", func() {
			// when:
			objectMeta := getObjectMetaFromSecretDefinition(sd)

			// then:
			Expect(objectMeta.Labels).Should(HaveKey("app.kubernetes.io/managed-by"))
			Expect(objectMeta.Annotations).Should(HaveKey("secrets-manager.tuenti.io/lastUpdateTime"))
		})
		It("getObjectMetaFromSecretDefinition should skip expected annotations", func() {
			// when:
			objectMeta := getObjectMetaFromSecretDefinition(sdWithSkipAnnotations)

			// then:
			Expect(objectMeta.Labels).Should(Not(HaveKey(corev1.LastAppliedConfigAnnotation)))
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

			// Stream generates values with DoSomething and sends them to out
			// until DoSomething returns an error or ctx.Done is closed.
			ctx, cancelfunc := context.WithCancel(context.Background())

			go func() {
				defer GinkgoRecover()
				Expect(mgr.Start(ctx)).NotTo(HaveOccurred())
				close(done)
			}()

			r2.Create(ctx, sdWatched)
			// Sleep for 4 * the reconcile interval set on the controller (just to be safe)
			time.Sleep(4 * time.Second)
			data, err := r2.getCurrentState(ctx, "watched", sdWatched.Spec.Name)
			Expect(err).To(BeNil())
			Expect(data).To(Equal(map[string][]byte{"watched": decodedBytes}))
			cancelfunc()

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

			ctx, cancelfunc := context.WithCancel(context.Background())
			go func() {
				defer GinkgoRecover()
				Expect(mgr.Start(ctx)).NotTo(HaveOccurred())
				close(done)
			}()

			r2.Create(ctx, sdNotWatched)
			// Sleep for 4 * the reconcile interval set on the controller (just to be safe)
			time.Sleep(4 * time.Second)
			data, err := r2.getCurrentState(ctx, "notwatched", sdNotWatched.Spec.Name)
			Expect(err.Error()).To(Equal("secrets \"secret-notwatched\" not found"))
			Expect(data).To(BeEmpty())
			cancelfunc()

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

			ctx, cancelfunc := context.WithCancel(context.Background())
			go func() {
				defer GinkgoRecover()
				Expect(mgr.Start(ctx)).NotTo(HaveOccurred())

				close(done)
			}()

			r2.Create(ctx, sdMultiWatched1)
			r2.Create(ctx, sdMultiWatched2)
			// Sleep for 4 * the reconcile interval set on the controller (just to be safe)
			time.Sleep(4 * time.Second)
			data, err2 := r2.getCurrentState(ctx, "watched1", sdMultiWatched1.Spec.Name)
			Expect(err2).To(BeNil())
			Expect(data).To(Equal(map[string][]byte{"multival1": decodedBytes}))

			data2, err3 := r2.getCurrentState(ctx, "watched2", sdMultiWatched2.Spec.Name)
			Expect(err3).To(BeNil())
			Expect(data2).To(Equal(map[string][]byte{"multival2": decodedBytes}))

			cancelfunc()

		}, 10)
	})
})
