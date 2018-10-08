package testutils

import (
	"github.com/tuenti/secrets-manager/kubernetes"
	"math/rand"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func NewFakeSecret(namespace string, name string) *kubernetes.Secret {
	return &kubernetes.Secret{
		Type:      "Opaque",
		Name:      name,
		Namespace: namespace,
		Data: map[string][]byte{
			"value": []byte(RandString(10)),
		},
	}
}

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
