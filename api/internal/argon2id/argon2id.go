// Package argon2id contains utilities for the argon2id protocol.
package argon2id

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

const (
	DefaultMemory      = 64 * 1024 // 64 MB
	DefaultIterations  = 1
	DefaultParallelism = 4
	DefaultSaltLength  = 16
	DefaultKeyLength   = 32
)

const (
	numHashSections = 6
)

type ArgonParams struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

var DefaultParams = ArgonParams{
	Memory:      DefaultMemory,
	Iterations:  DefaultIterations,
	Parallelism: DefaultIterations,
	SaltLength:  DefaultSaltLength,
	KeyLength:   DefaultKeyLength,
}

// HashAndEncode encodes a message with the given parameters.
func HashAndEncode(message string, p ArgonParams) (string, error) {
	salt := make([]byte, p.SaltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("creating random salt: %w", err)
	}
	return encodeHashWithSalt(message, p, salt), nil
}

// encodeHashWithSalt encodes a message with the given parameters and salt.
func encodeHashWithSalt(message string, p ArgonParams, salt []byte) string {
	b64Hash := base64.RawStdEncoding.EncodeToString(HashWithSalt(message, p, salt))
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.Memory, p.Iterations,
		p.Parallelism, b64Salt, b64Hash)

	return encodedHash
}

// HashWithSalt hashes a message with the given parametesr and salt.
// This function is typicallly used when constructing a hash for comparison
// purposes.
func HashWithSalt(message string, p ArgonParams, salt []byte) []byte {
	return argon2.IDKey(
		[]byte(message), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)
}

// DecodeHash decodes a argon2id encoded hash.
func DecodeHash(encodedHash string) (p *ArgonParams, salt []byte, hash []byte, err error) {
	argonVals := strings.Split(encodedHash, "$")
	if len(argonVals) != numHashSections {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	var version int
	_, err = fmt.Sscanf(argonVals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	p = &ArgonParams{}
	_, err = fmt.Sscanf(argonVals[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(argonVals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.SaltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(argonVals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.KeyLength = uint32(len(hash))
	return p, salt, hash, nil
}
