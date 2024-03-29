package errors

import "fmt"

// Error Types constants
const (
	UnknownErrorType                   = "UnknownError"
	BackendNotImplementedErrorType     = "BackendNotImplementedError"
	BackendSecretNotFoundErrorType     = "BackendSecretNotFoundError"
	BackendSecretForbiddenErrorType    = "BackendSecretForbiddenError"
	K8sSecretNotFoundErrorType         = "K8sSecretNotFoundError"
	EncodingNotImplementedErrorType    = "EncodingNotImplementedError"
	VaultEngineNotImplementedErrorType = "VaultEngineNotImplementedError"
	VaultTokenNotRenewableErrorType    = "VaultTokenNotRenewableError"
)

// BackendNotImplementedError will be raised if the selected backend is not implemented
type BackendNotImplementedError struct {
	ErrType string
	Backend string
}

// BackendSecretNotFoundError will be raised if secret is not found in the selected backend
type BackendSecretNotFoundError struct {
	ErrType string
	Path    string
	Key     string
}

// K8sSecretNotFoundError will be raised if secret is not found by its name in the given namespace
type K8sSecretNotFoundError struct {
	ErrType   string
	Name      string
	Namespace string
}

// EncodingNotImplementedError will be raised if the selected encoding is not implemented
type EncodingNotImplementedError struct {
	ErrType  string
	Encoding string
}

// VaultEngineNotImplementedError will be raised if the selected engine is not implemented
type VaultEngineNotImplementedError struct {
	ErrType string
	Engine  string
}

// VaultTokenNotRenewableError will be raised if secrets-manager Vault token is not renewable
type VaultTokenNotRenewableError struct {
	ErrType string
}

func getErrorType(err error) string {
	switch err.(type) {
	case *BackendNotImplementedError:
		return BackendNotImplementedErrorType
	case *BackendSecretNotFoundError:
		return BackendSecretNotFoundErrorType
	case *K8sSecretNotFoundError:
		return K8sSecretNotFoundErrorType
	case *EncodingNotImplementedError:
		return EncodingNotImplementedErrorType
	case *VaultEngineNotImplementedError:
		return VaultEngineNotImplementedErrorType
	case *VaultTokenNotRenewableError:
		return VaultTokenNotRenewableErrorType
	default:
		return UnknownErrorType
	}
}

func (e BackendNotImplementedError) Error() string {
	return fmt.Sprintf("[%s] backend %s not supported", e.ErrType, e.Backend)
}

func (e BackendSecretNotFoundError) Error() string {
	return fmt.Sprintf("[%s] secret key %s not found at %s", e.ErrType, e.Key, e.Path)
}

func (e K8sSecretNotFoundError) Error() string {
	return fmt.Sprintf("[%s] secret '%s/%s' not found", e.ErrType, e.Namespace, e.Name)
}

func (e EncodingNotImplementedError) Error() string {
	return fmt.Sprintf("[%s] encoding %s not supported", e.ErrType, e.Encoding)
}

func (e VaultEngineNotImplementedError) Error() string {
	return fmt.Sprintf("[%s] vault engine %s not supported", e.ErrType, e.Engine)
}

func (e VaultTokenNotRenewableError) Error() string {
	return fmt.Sprintf("[%s] vault token not renewable", e.ErrType)
}

// IsBackendNotImplemented returns true if the error is type of BackendNotImplementedError and false otherwise
func IsBackendNotImplemented(err error) bool {
	return getErrorType(err) == BackendNotImplementedErrorType
}

// IsBackendSecretNotFound returns true if the error is type of BackendSecretNotFound and false otherwise
func IsBackendSecretNotFound(err error) bool {
	return getErrorType(err) == BackendSecretNotFoundErrorType
}

// IsK8sSecretNotFound returns true if the error is type of K8sSecretNotFound and false otherwise
func IsK8sSecretNotFound(err error) bool {
	return getErrorType(err) == K8sSecretNotFoundErrorType
}

// IsEncodingNotImplemented returns true if the error is type of EncodingNotImplementedError and false otherwise
func IsEncodingNotImplemented(err error) bool {
	return getErrorType(err) == EncodingNotImplementedErrorType
}

// IsVaultEngineNotImplemented returns true if the error is type of VaultEngineNotImplementedError and false otherwise
func IsVaultEngineNotImplemented(err error) bool {
	return getErrorType(err) == VaultEngineNotImplementedErrorType
}

// IsVaultTokenNotRenewable returns true if the error is type of VaultTokenNotRenewableError and false otherwise
func IsVaultTokenNotRenewable(err error) bool {
	return getErrorType(err) == VaultTokenNotRenewableErrorType
}
