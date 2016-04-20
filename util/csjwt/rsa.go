package csjwt

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
)

// SigningMethodRSA implements the RSA family of signing methods signing methods
type SigningMethodRSA struct {
	Name string
	Hash crypto.Hash
}

func newSigningMethodRSA(n string, h crypto.Hash) *SigningMethodRSA {
	return &SigningMethodRSA{Name: n, Hash: h}
}

// NewSigningMethodRS256 creates a new 256bit RSA SHA instance and registers it.
func NewSigningMethodRS256() *SigningMethodRSA {
	return newSigningMethodRSA(RS256, crypto.SHA256)
}

// NewSigningMethodRS384 creates a new 384bit RSA SHA instance and registers it.
func NewSigningMethodRS384() *SigningMethodRSA {
	return newSigningMethodRSA(RS384, crypto.SHA384)
}

// NewSigningMethodRS512 creates a new 512bit RSA SHA instance and registers it.
func NewSigningMethodRS512() *SigningMethodRSA {
	return newSigningMethodRSA(RS512, crypto.SHA512)
}

func (m *SigningMethodRSA) Alg() string {
	return m.Name
}

// Verify implements the Verify method from SigningMethod interface.
// For the key you can use any of the WithRSA*Key*() functions.
func (m *SigningMethodRSA) Verify(signingString, signature []byte, key Key) error {
	if key.Error != nil {
		return key.Error
	}
	if key.rsaKeyPub == nil {
		return errRSAPublicKeyEmpty
	}

	// Decode the signature
	sig, err := DecodeSegment(signature)
	if err != nil {
		return err
	}

	// Create hasher
	if !m.Hash.Available() {
		return errRSAHashUnavailable
	}
	hasher := m.Hash.New()
	if _, err := hasher.Write(signingString); err != nil {
		return err
	}

	// Verify the signature
	return rsa.VerifyPKCS1v15(key.rsaKeyPub, m.Hash, hasher.Sum(nil), sig)
}

// Sign implements the Sign method from SigningMethod interface.
// For the key you can use any of the WithRSAPrivateKey*() functions.
func (m *SigningMethodRSA) Sign(signingString []byte, key Key) ([]byte, error) {
	if key.Error != nil {
		return nil, key.Error
	}
	if key.rsaKeyPriv == nil {
		return nil, errRSAPrivateKeyEmpty
	}

	// Create the hasher
	if !m.Hash.Available() {
		return nil, errRSAHashUnavailable
	}

	hasher := m.Hash.New()
	if _, err := hasher.Write(signingString); err != nil {
		return nil, err
	}

	// Sign the string and return the encoded bytes
	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, key.rsaKeyPriv, m.Hash, hasher.Sum(nil))
	if err != nil {
		return nil, err
	}
	return EncodeSegment(sigBytes), nil
}