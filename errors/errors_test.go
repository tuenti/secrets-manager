package errors

import (
	e "errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorString(t *testing.T) {
	err1 := &BackendNotImplementedError{ErrType: BackendNotImplementedErrorType, Backend: "foo"}
	assert.EqualError(t, err1, fmt.Sprintf("[%s] backend %s not supported", err1.ErrType, err1.Backend))
	err2 := &BackendSecretNotFoundError{ErrType: BackendSecretNotFoundErrorType, Path: "foo", Key: "bar"}
	assert.EqualError(t, err2, fmt.Sprintf("[%s] secret key %s not found at %s", err2.ErrType, err2.Key, err2.Path))
	err3 := &K8sSecretNotFoundError{ErrType: K8sSecretNotFoundErrorType, Name: "foo", Namespace: "bar"}
	assert.EqualError(t, err3, fmt.Sprintf("[%s] secret '%s/%s' not found", err3.ErrType, err3.Namespace, err3.Name))
	err5 := &EncodingNotImplementedError{ErrType: EncodingNotImplementedErrorType, Encoding: "foo"}
	assert.EqualError(t, err5, fmt.Sprintf("[%s] encoding %s not supported", err5.ErrType, err5.Encoding))
	err6 := &VaultEngineNotImplementedError{ErrType: VaultEngineNotImplementedErrorType, Engine: "foo"}
	assert.EqualError(t, err6, fmt.Sprintf("[%s] vault engine %s not supported", err6.ErrType, err6.Engine))
	err7 := &VaultTokenNotRenewableError{ErrType: VaultTokenNotRenewableErrorType}
	assert.EqualError(t, err7, fmt.Sprintf("[%s] vault token not renewable", err7.ErrType))
}

func TestGetErrorType(t *testing.T) {
	err1 := e.New("foo")
	assert.Equal(t, getErrorType(err1), UnknownErrorType)
	err2 := &BackendNotImplementedError{ErrType: BackendNotImplementedErrorType}
	assert.Equal(t, getErrorType(err2), BackendNotImplementedErrorType)
	err3 := &BackendSecretNotFoundError{ErrType: BackendSecretNotFoundErrorType}
	assert.Equal(t, getErrorType(err3), BackendSecretNotFoundErrorType)
	err4 := &K8sSecretNotFoundError{ErrType: K8sSecretNotFoundErrorType}
	assert.Equal(t, getErrorType(err4), K8sSecretNotFoundErrorType)
	err6 := &EncodingNotImplementedError{ErrType: EncodingNotImplementedErrorType}
	assert.Equal(t, getErrorType(err6), EncodingNotImplementedErrorType)
	err7 := &VaultEngineNotImplementedError{ErrType: VaultEngineNotImplementedErrorType}
	assert.Equal(t, getErrorType(err7), VaultEngineNotImplementedErrorType)
	err8 := &VaultTokenNotRenewableError{ErrType: VaultTokenNotRenewableErrorType}
	assert.Equal(t, getErrorType(err8), VaultTokenNotRenewableErrorType)
}

func TestIsBackendNotImplemented(t *testing.T) {
	err := &BackendNotImplementedError{ErrType: BackendNotImplementedErrorType}
	assert.True(t, IsBackendNotImplemented(err))
	err2 := e.New("foo")
	assert.False(t, IsBackendNotImplemented(err2))
}

func TestIsBackendSecretNotFound(t *testing.T) {
	err := &BackendSecretNotFoundError{ErrType: BackendSecretNotFoundErrorType}
	assert.True(t, IsBackendSecretNotFound(err))
	err2 := e.New("foo")
	assert.False(t, IsBackendSecretNotFound(err2))
}

func TestIsK8sSecretNotFound(t *testing.T) {
	err := &K8sSecretNotFoundError{ErrType: K8sSecretNotFoundErrorType}
	assert.True(t, IsK8sSecretNotFound(err))
	err2 := e.New("foo")
	assert.False(t, IsK8sSecretNotFound(err2))
}

func TestIsEncodingNotImplemented(t *testing.T) {
	err := &EncodingNotImplementedError{ErrType: EncodingNotImplementedErrorType}
	assert.True(t, IsEncodingNotImplemented(err))
	err2 := e.New("foo")
	assert.False(t, IsEncodingNotImplemented(err2))
}

func TestIsVaultEngineNotImplemented(t *testing.T) {
	err := &VaultEngineNotImplementedError{ErrType: VaultEngineNotImplementedErrorType}
	assert.True(t, IsVaultEngineNotImplemented(err))
}

func TestIsVaultTokenNotRenewable(t *testing.T) {
	err := &VaultTokenNotRenewableError{ErrType: VaultTokenNotRenewableErrorType}
	assert.True(t, IsVaultTokenNotRenewable(err))
}
