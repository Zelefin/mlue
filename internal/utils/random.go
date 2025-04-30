package utils

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

func GenerateRandomState() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal("failed to generate random state:", err)
	}
	return base64.URLEncoding.EncodeToString(b)
}
