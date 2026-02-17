// Package crypto provides cryptographic utilities for backup encryption.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// DeriveBackupKey derives an encryption key bound to a specific application.
// It applies HMAC-SHA256(pepper, password) before Argon2id, ensuring that
// even with the correct password, the key cannot be reproduced without the
// application-specific pepper compiled into the binary.
func DeriveBackupKey(password string, pepper, salt []byte, params *Argon2Params) []byte {
	mac := hmac.New(sha256.New, pepper)
	mac.Write([]byte(password))
	pepperedPassword := mac.Sum(nil)

	if params == nil {
		params = DefaultArgon2Params
	}
	return argon2.IDKey(
		pepperedPassword,
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)
}

// Argon2Params holds parameters for Argon2id key derivation.
type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	KeyLength   uint32
}

// DefaultArgon2Params provides safe defaults for key derivation.
var DefaultArgon2Params = &Argon2Params{
	Memory:      64 * 1024,
	Iterations:  3,
	Parallelism: 2,
	KeyLength:   32, // 256-bit key for AES-256
}

const (
	// NonceSize is the standard AES-GCM nonce length.
	NonceSize = 12
	// SaltSize is the Argon2 salt length.
	SaltSize = 32
)

var (
	ErrDecryptionFailed = errors.New("decryption failed: invalid password or corrupted data")
)

// GenerateSalt creates a cryptographically secure random salt.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generating salt: %w", err)
	}
	return salt, nil
}

// DeriveKey stretches a password using Argon2id and returns a key
// suitable for AES-256-GCM.
func DeriveKey(password string, salt []byte, params *Argon2Params) []byte {
	if params == nil {
		params = DefaultArgon2Params
	}
	return argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)
}

// Encrypt encrypts plaintext using AES-256-GCM with the given key.
// Returns the nonce and ciphertext separately so they can be stored
// in the archive header.
func Encrypt(plaintext, key []byte) (nonce []byte, ciphertext []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("creating GCM: %w", err)
	}

	nonce = make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("generating nonce: %w", err)
	}

	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return nonce, ciphertext, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM with the given key and nonce.
func Decrypt(ciphertext, key, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}
