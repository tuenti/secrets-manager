package secretsmanager

import (
	"fmt"
	"reflect"

	gomock "github.com/golang/mock/gomock"
	"github.com/tuenti/secrets-manager/kubernetes"
)

type secretMatcher struct {
	secret *kubernetes.Secret
}

func (s secretMatcher) Matches(x interface{}) bool {
	inputSecret, ok := x.(*kubernetes.Secret)
	if !ok {
		return false
	}
	if inputSecret.Name != s.secret.Name {
		return false
	}
	if inputSecret.Namespace != s.secret.Namespace {
		return false
	}
	if !reflect.DeepEqual(inputSecret.Data, s.secret.Data) {
		return false
	}
	return true
}

func (s secretMatcher) String() string {
	return fmt.Sprintf("is same secret (ignoring labels): %v\n", s.secret)
}

func EqSecret(secret *kubernetes.Secret) gomock.Matcher {
	return secretMatcher{secret}
}
