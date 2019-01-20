package errors

import (
	e "errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetErrorType(t *testing.T) {
	err1 := e.New("foo")
	assert.Equal(t, getErrorType(err1), UnknownErrorType)
	err2 := &BackendNotImplementedError{ErrType: BackendNotImplementedErrorType}
	assert.Equal(t, getErrorType(err2), BackendNotImplementedErrorType)
	err3 := &BackendSecretNotFoundError{ErrType: BackendSecretNotFoundErrorType}
	assert.Equal(t, getErrorType(err3), BackendSecretNotFoundErrorType)
	err4 := &K8sSecretNotFoundError{ErrType: K8sSecretNotFoundErrorType}
	assert.Equal(t, getErrorType(err4), K8sSecretNotFoundErrorType)
	err5 := &InvalidConfigmapNameError{ErrType: InvalidConfigmapNameErrorType}
	assert.Equal(t, getErrorType(err5), InvalidConfigmapNameErrorType)
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

func TestIsInvalidConfigmapName(t *testing.T) {
	err := &InvalidConfigmapNameError{ErrType: InvalidConfigmapNameErrorType}
	assert.True(t, IsInvalidConfigmapName(err))
	err2 := e.New("foo")
	assert.False(t, IsInvalidConfigmapName(err2))
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
