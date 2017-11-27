package berry

import (
	"encoding/binary"
	"encoding/hex"
	"sync/atomic"
)

func generateID() string {
	n := atomic.AddUint32(&idcounter, 1)

	var rawid [4]byte
	binary.BigEndian.PutUint32(rawid[:], n)
	return hex.EncodeToString(rawid[:])
}

var idcounter = uint32(0)
