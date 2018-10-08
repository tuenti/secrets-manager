package backend

import (
	"encoding/base64"

	"github.com/tuenti/secrets-manager/errors"
)

const (
	// Base64EncodingType is the internal code to represent a base64 encoding
	Base64EncodingType = "base64"

	// TextEncodingType is the internal code to represent a basic text encoding
	TextEncodingType = "text"

	// DefaultEncodingType is the default encoding to use.
	DefaultEncodingType = "text"
)

// Decoder interface represents anything that can implement DecodeString: get some bytes from input string
type Decoder interface {
	DecodeString(input string) ([]byte, error)
}

// Base64Decoder represents a Decoder for base64 text
type Base64Decoder struct {
	Encoding string
}

// TextDecoder represents a Decoder for plain text
type TextDecoder struct {
	Encoding string
}

// DecodeString for Base64Decoder will get the text version (in bytes) of the input base64 text
func (d Base64Decoder) DecodeString(input string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil, err
	}
	return data, err
}

// DecodeString for TextDecoder, will simply cast to []bytes the input text
func (d TextDecoder) DecodeString(input string) ([]byte, error) {
	return []byte(input), nil
}

// NewDecoder returns a new Decoder implementation or an error if the provided encoding is not implemented
func NewDecoder(encoding string) (Decoder, error) {
	if encoding == "" {
		encoding = DefaultEncodingType
	}
	switch encoding {
	case Base64EncodingType:
		return Base64Decoder{Encoding: encoding}, nil
	case TextEncodingType:
		return TextDecoder{Encoding: encoding}, nil
	default:
		return nil, &errors.EncodingNotImplementedError{ErrType: errors.EncodingNotImplementedErrorType, Encoding: encoding}
	}
}
