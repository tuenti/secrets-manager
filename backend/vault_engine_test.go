package backend

import (
	"fmt"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/tuenti/secrets-manager/errors"
)

func TestNewEngineKV1(t *testing.T) {
	eng := "kv1"
	engine, err := newEngine(eng)
	assert.Nil(t, err)
	assert.Equal(t, eng, engine.(kvEngineV1).name)
}

func TestNewEngineKV2(t *testing.T) {
	eng := "kv2"
	engine, err := newEngine(eng)
	assert.Nil(t, err)
	assert.Equal(t, eng, engine.(kvEngineV2).name)
}

func TestNotImplementedEngine(t *testing.T) {
	eng := "kv3"
	_, err := newEngine(eng)
	assert.NotNil(t, err)
	assert.EqualError(t, err, fmt.Sprintf("[%s] vault engine %s not supported", errors.VaultEngineNotImplementedErrorType, eng))
}

func TestGetDataKv1(t *testing.T) {
	data := make(map[string]interface{})
	data["foo"] = "bar"
	s := &api.Secret{Data: data}
	engine, _ := newEngine("kv1")
	d := engine.getData(s)
	assert.NotNil(t, d)
	assert.Equal(t, data, d)
}

func TestGetDataKv2(t *testing.T) {
	data := make(map[string]interface{})
	nested := make(map[string]interface{})
	nested["foo"] = "bar"
	data["data"] = nested
	s := &api.Secret{Data: data}
	engine, _ := newEngine("kv2")
	d := engine.getData(s)
	assert.NotNil(t, d)
	assert.Equal(t, nested, d)
}

func TestGetDataKv2WithKv1Engine(t *testing.T) {
	data := make(map[string]interface{})
	nested := make(map[string]interface{})
	nested["foo"] = "bar"
	data["data"] = nested
	s := &api.Secret{Data: data}
	engine, _ := newEngine("kv1")
	d := engine.getData(s)
	assert.NotNil(t, d)
	assert.Equal(t, data, d)
}
