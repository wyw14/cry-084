package identity

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/local/cry-084/internal/domain/shared"
)

type Random struct{}

func (Random) New() shared.ID {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		panic(err)
	}
	return shared.ID(hex.EncodeToString(raw[:]))
}
