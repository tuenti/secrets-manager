package controllers

import (
	"context"
	"encoding/base64"

	//"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	secretsmanagerv1alpha1 "github.com/tuenti/secrets-manager/api/v1alpha1"
	"github.com/tuenti/secrets-manager/errors"

	//corev1 "k8s.io/api/core/v1"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	encodedValue = "bG9yZW0gaXBzdW0gZG9ybWEK"
	decodedValue = "lorem ipsum dorma"
)

var _ = Describe("SecretsManager", func() {
	var (
		r  *SecretDefinitionReconciler
		sd = &secretsmanagerv1alpha1.SecretDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "secretdef-test",
			},
			Spec: secretsmanagerv1alpha1.SecretDefinitionSpec{
				Name: "secret-test",
				Type: "Opaque",
				KeysMap: map[string]secretsmanagerv1alpha1.DataSource{
					"foo": secretsmanagerv1alpha1.DataSource{
						Path:     "secret/data/pathtosecret1",
						Key:      "value",
						Encoding: "base64",
					},
				},
			},
		}
		sd2 = &secretsmanagerv1alpha1.SecretDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "secretdef-test2",
			},
			Spec: secretsmanagerv1alpha1.SecretDefinitionSpec{
				Name: "secret-test2",
				Type: "Opaque",
				KeysMap: map[string]secretsmanagerv1alpha1.DataSource{
					"foo2": secretsmanagerv1alpha1.DataSource{
						Path:     "secret/data/pathtosecret1",
						Key:      "value",
						Encoding: "base64",
					},
				},
			},
		}
		sdBackendSecretNotFound = &secretsmanagerv1alpha1.SecretDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "secretdef-beckend-secret-not-found",
			},
			Spec: secretsmanagerv1alpha1.SecretDefinitionSpec{
				Name: "secret-backend-secret-not-found",
				Type: "Opaque",
				KeysMap: map[string]secretsmanagerv1alpha1.DataSource{
					"foo3": secretsmanagerv1alpha1.DataSource{
						Path:     "secret/data/notfound",
						Key:      "value",
						Encoding: "base64",
					},
				},
			},
		}
		sdWrongEncoding = &secretsmanagerv1alpha1.SecretDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "secretdef-wrong-encoding",
			},
			Spec: secretsmanagerv1alpha1.SecretDefinitionSpec{
				Name: "secret-wrong-encoding",
				Type: "Opaque",
				KeysMap: map[string]secretsmanagerv1alpha1.DataSource{
					"foo4": secretsmanagerv1alpha1.DataSource{
						Path:     "secret/data/pathtosecret1",
						Key:      "value",
						Encoding: "base65",
					},
				},
			},
		}
	)

	BeforeEach(func() {
		r = getReconciler()
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
})
