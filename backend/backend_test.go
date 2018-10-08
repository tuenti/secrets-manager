package backend

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tuenti/secrets-manager/errors"
)

func TestNotImplementedBackend(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := Config{}
	backend := "foo"
	_, err := NewBackendClient(ctx, backend, nil, cfg)
	assert.EqualError(t, err, fmt.Sprintf("[%s] backend %s not supported", errors.BackendNotImplementedErrorType, backend))
}
