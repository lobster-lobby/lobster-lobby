package services

import (
	"crypto/rand"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

const (
	apiKeyPrefix  = "ll_"
	apiKeyLength  = 48
	keyPrefixLen  = 8
	bcryptCost    = 10
	apiKeyCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type APIKeyService struct{}

func NewAPIKeyService() *APIKeyService {
	return &APIKeyService{}
}

// GenerateKey creates a new API key with format ll_ + 48 random alphanumeric chars.
// Returns the full key, the prefix (first 8 chars after ll_), and the bcrypt hash.
func (s *APIKeyService) GenerateKey() (fullKey, prefix, hash string, err error) {
	randomPart, err := randomAlphanumeric(apiKeyLength)
	if err != nil {
		return "", "", "", err
	}

	fullKey = apiKeyPrefix + randomPart
	prefix = randomPart[:keyPrefixLen]

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(fullKey), bcryptCost)
	if err != nil {
		return "", "", "", err
	}
	hash = string(hashBytes)

	return fullKey, prefix, hash, nil
}

// VerifyKey checks a full API key against a bcrypt hash.
func (s *APIKeyService) VerifyKey(fullKey, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(fullKey)) == nil
}

// ExtractPrefix extracts the 8-char prefix from a full API key (after "ll_").
func (s *APIKeyService) ExtractPrefix(fullKey string) string {
	if len(fullKey) <= len(apiKeyPrefix)+keyPrefixLen {
		return ""
	}
	return fullKey[len(apiKeyPrefix) : len(apiKeyPrefix)+keyPrefixLen]
}

func randomAlphanumeric(n int) (string, error) {
	b := make([]byte, n)
	max := big.NewInt(int64(len(apiKeyCharset)))
	for i := range b {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = apiKeyCharset[idx.Int64()]
	}
	return string(b), nil
}
