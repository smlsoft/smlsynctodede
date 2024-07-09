package myglobal

import (
	"crypto/sha256"
	"encoding/hex"
)

var ShopId = "2QoilMQkX9i6vtAE88ilEubnrhz"

func CalculateSHA256(input string) string {
	hash := sha256.New()
	hash.Write([]byte(input))
	hashBytes := hash.Sum(nil)
	return hex.EncodeToString(hashBytes)
}
