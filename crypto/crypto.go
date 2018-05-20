// Package crypto includes all of the project crypto functions together for easy review. This package is
// derived substantially from https://github.com/gtank/cryptopasta, with thanks to George Tankersley @gtank__
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"github.com/pkg/errors"
	"golang.org/x/crypto/scrypt"
)

// Decrypt decrypts data using 256-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered.
func Decrypt(ciphertext []byte, key *[32]byte) (plaintext []byte, err error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	return gcm.Open(nil,
		ciphertext[:gcm.NonceSize()],
		ciphertext[gcm.NonceSize():],
		nil,
	)
}

// Decrypt64 accepts a string containing base64 encoded binary data and wraps Decrypt()
func Decrypt64(ciphertext string, key *[32]byte) (plaintext []byte, err error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode ciphertext as base64 string")
	}
	return Decrypt(data, key)
}

// Stretch uses scrypt to stretch the provided passphrase to something appropriate for use with Encrypt/Decrypt
// functions and hopefully more resilient. Parameters were selected by benchmark to use 64MiB of memory and > 0.5
// seconds to hash
func Stretch(passphrase string) (key *[32]byte, err error) {
	if len(passphrase) < 8 {
		return nil, errors.New("passphrase is too short")
	}
	keySlice, err := scrypt.Key([]byte(passphrase), []byte("unsalted"), 65536, 8, 4, 32)
	// scrypt only gives us back a slice so we explicitly check it before conversion
	if len(keySlice) != 32 {
		return nil, errors.Wrap(err, "invalid key returned")
	}
	copy(key[:], keySlice)
	return key, nil
}
