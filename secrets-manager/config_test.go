package secretsmanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSecretDefsFromYaml(t *testing.T) {
	configText := `
- name: supersecret1
  type: kubernetes.io/tls
  namespaces:
  - default
  data:
    tls.crt:
      path: secret/data/pathtosecret1
      key: value
    tls.key:
      path: secret/data/pathtosecret2
      key: value

- name: supersecret2
  type: Opaque
  namespaces:
  - default
  data:
    value1:
      path: secret/data/pathtosecret1
      key: value
    value2:
      path: secret/data/pathtosecret1
      key: value
`

	secretDefs, err := parseSecretDefsFromYaml(configText)

	assert.Nil(t, err)
	assert.Len(t, secretDefs, 2)
}

func TestParseSecretDefsFromYamlInvalidYaml(t *testing.T) {
	configText := `
- something: that
  doesnt: match
  the: exptected
   	yaml: structure 
`

	_, err := parseSecretDefsFromYaml(configText)

	assert.NotNil(t, err)
}
