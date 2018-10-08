package backend

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tuenti/secrets-manager/errors"
)

func TestNotImplementedDecoder(t *testing.T) {
	encoding := "foo"
	_, err := NewDecoder(encoding)
	assert.EqualError(t, err, fmt.Sprintf("[%s] encoding %s not supported", errors.EncodingNotImplementedErrorType, encoding))
}

func TestGetB64Decoder(t *testing.T) {
	encoding := "base64"
	decoder, err := NewDecoder(encoding)
	b64Decoder := decoder.(Base64Decoder)
	assert.Nil(t, err)
	assert.Equal(t, encoding, b64Decoder.Encoding)
}

func TestGetTextDecoder(t *testing.T) {
	encoding := "text"
	decoder, err := NewDecoder(encoding)
	textDecoder := decoder.(TextDecoder)
	assert.Nil(t, err)
	assert.Equal(t, encoding, textDecoder.Encoding)
}

func TestGetTextDecoderFromEmptyString(t *testing.T) {
	encoding := ""
	decoder, err := NewDecoder(encoding)
	textDecoder := decoder.(TextDecoder)
	assert.Nil(t, err)
	assert.Equal(t, encoding, textDecoder.Encoding)
}

func TestDecodeB64String(t *testing.T) {
	b64data := "dGVzdGluZyBiYXNlNjQgZGVjb2RpbmcK"
	decoder, err := NewDecoder("base64")
	data, err := decoder.DecodeString(b64data)
	assert.Nil(t, err)
	assert.Equal(t, "testing base64 decoding", fmt.Sprintf("%s", data))
}

func TestDecodeInvalidB64String(t *testing.T) {
	b64data := "Invalid b64 data"
	decoder, err := NewDecoder("base64")
	data, err := decoder.DecodeString(b64data)
	assert.NotNil(t, err)
	assert.Nil(t, data)
}

func TestDecodeText(t *testing.T) {
	text := "secret text"
	decoder, err := NewDecoder("text")
	data, err := decoder.DecodeString(text)
	assert.Nil(t, err)
	assert.Equal(t, text, fmt.Sprintf("%s", data))
}
