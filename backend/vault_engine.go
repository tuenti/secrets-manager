package backend

import (
	"github.com/hashicorp/vault/api"
	"github.com/tuenti/secrets-manager/errors"
)

const (
	kvEngineV1Name = "kv1"
	kvEngineV2Name = "kv2"
)

type engine interface {
	getData(s *api.Secret) map[string]interface{}
}

type kvEngineV1 struct {
	name string
}

type kvEngineV2 struct {
	name string
}

func (e kvEngineV1) getData(s *api.Secret) map[string]interface{} {
	return s.Data
}

func (e kvEngineV2) getData(s *api.Secret) map[string]interface{} {
	if s.Data["data"] == nil {
		return nil
	}
	return s.Data["data"].(map[string]interface{})
}

func newEngine(eng string) (engine, error) {
	if eng == "" {
		eng = kvEngineV2Name
	}
	switch eng {
	case kvEngineV1Name:
		return kvEngineV1{name: kvEngineV1Name}, nil
	case kvEngineV2Name:
		return kvEngineV2{name: kvEngineV2Name}, nil
	default:
		return nil, &errors.VaultEngineNotImplementedError{ErrType: errors.VaultEngineNotImplementedErrorType, Engine: eng}
	}
}
