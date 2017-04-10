package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
)

// ParsePublicKey parses a public PEM file into a proper
// PublicKey struct
func ParsePublicKey(pubPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		err := errors.New("Failed to parse PEM block containing the public key")
		return nil, err
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	return pubKey.(*rsa.PublicKey), err
}

// VerifySignature verifies the signature sent with the request with
// the pem we retried from the Chef server
func VerifySignature(encodedSig, pubPEM string) error {
	pubKey, err := ParsePublicKey(pubPEM)
	if err != nil {
		return err
	}
	hashed := sha512.Sum512([]byte(pubPEM))
	signature, err := base64.StdEncoding.DecodeString(encodedSig)
	if err != nil {
		return err
	}
	err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA512, hashed[:], signature)
	return err
}
