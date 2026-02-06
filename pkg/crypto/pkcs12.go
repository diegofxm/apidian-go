package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"golang.org/x/crypto/pkcs12"
)

// ValidatePKCS12 validates a PKCS12 certificate with the given password
// Returns error if the certificate is invalid or password is incorrect
func ValidatePKCS12(certificateData []byte, password string) error {
	// Try to decode the PKCS12 file with the provided password
	_, _, err := pkcs12.Decode(certificateData, password)
	if err != nil {
		return fmt.Errorf("invalid certificate or password: %w", err)
	}

	return nil
}

// DecodePKCS12 decodes a base64 encoded PKCS12 certificate
func DecodePKCS12(base64Cert string) ([]byte, error) {
	certificateData, err := base64.StdEncoding.DecodeString(base64Cert)
	if err != nil {
		return nil, errors.New("certificate must be valid base64 encoded data")
	}

	return certificateData, nil
}

// EncryptPassword encrypts a password using AES-256-GCM with the ENCRYPTION_KEY from environment
// This provides strong encryption with authentication (AEAD)
func EncryptPassword(password string) (string, error) {
	// Get encryption key from environment
	keyHex := os.Getenv("ENCRYPTION_KEY")
	if keyHex == "" {
		return "", errors.New("ENCRYPTION_KEY not set in environment")
	}

	// Decode hex key to bytes (must be 32 bytes for AES-256)
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", fmt.Errorf("invalid ENCRYPTION_KEY format: %w", err)
	}

	if len(key) != 32 {
		return "", fmt.Errorf("ENCRYPTION_KEY must be 32 bytes (64 hex chars), got %d bytes", len(key))
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt password
	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)

	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptPassword decrypts a password encrypted with EncryptPassword using AES-256-GCM
func DecryptPassword(encryptedPassword string) (string, error) {
	// Get encryption key from environment
	keyHex := os.Getenv("ENCRYPTION_KEY")
	if keyHex == "" {
		return "", errors.New("ENCRYPTION_KEY not set in environment")
	}

	// Decode hex key to bytes
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", fmt.Errorf("invalid ENCRYPTION_KEY format: %w", err)
	}

	if len(key) != 32 {
		return "", fmt.Errorf("ENCRYPTION_KEY must be 32 bytes (64 hex chars), got %d bytes", len(key))
	}

	// Decode base64 ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", fmt.Errorf("invalid encrypted password format: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}
