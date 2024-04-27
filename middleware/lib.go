package middleware

import (
	"backend/config"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"time"
)

var timeoutLimit = 30 * time.Minute
var employeeCollection = config.GetCollection(config.ConnectDB(), "employee")

/*
Decrypt the token from the client

params: ciphertext []byte The encrypted token

key []byte The key to decrypt the token

return: []byte The decrypted token

error The error if the decryption fails
*/
func Decrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return aesGCM.Open(nil, nonce, ciphertext, nil)
}
