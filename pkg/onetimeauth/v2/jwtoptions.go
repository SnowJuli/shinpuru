package onetimeauth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// DefaultOptions holds default values
// for the OneTimeAuth configuration.
var DefaultOptions = &JwtOptions{
	Issuer:           "generic issuer",
	Lifetime:         time.Minute,
	SigningKeyLength: 128,
	TokenKeyLength:   32,
	SigningMethod:    jwt.SigningMethodHS256,
}

// JwtOptions holds configuration parameters
// for the JwtOneTimeAuth instance.
type JwtOptions struct {
	Issuer           string        `json:"issuer"`
	Lifetime         time.Duration `json:"duration"`
	SigningKeyLength int           `json:"signing_key_length"`
	TokenKeyLength   int           `json:"token_key_length"`
	SigningMethod    jwt.SigningMethod
}

func (o *JwtOptions) complete() {
	if o.Issuer == "" {
		o.Issuer = DefaultOptions.Issuer
	}
	if o.Lifetime <= 0 {
		o.Lifetime = DefaultOptions.Lifetime
	}
	if o.SigningKeyLength <= 0 {
		o.SigningKeyLength = DefaultOptions.SigningKeyLength
	}
	if o.TokenKeyLength <= 0 {
		o.TokenKeyLength = DefaultOptions.TokenKeyLength
	}
	if o.SigningMethod == nil {
		o.SigningMethod = DefaultOptions.SigningMethod
	}
}
