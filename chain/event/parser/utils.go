package event

import (
	"crypto/sha256"
)

func EventDiscriminator(name string) []byte {
	return signHash("event", name)
}

func signHash(namespace string, name string) []byte {
	data := namespace + ":" + name
	sum := sha256.Sum256([]byte(data))
	return sum[0:8]
}
