package berry

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"math"
	"time"
)

func generateID() string {
	return <-readyid
}

var readyid = make(chan string, 32)

func init() {
	go func() {
		var (
			counter  uint16
			failures int
		)
		raw := make([]byte, 4)

		for {
			if _, err := rand.Read(raw[:2]); err != nil {
				if failures > 10 {
					panic(err)
				}
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if counter == math.MaxUint16 {
				counter = 0
			}
			counter++

			binary.BigEndian.PutUint16(raw[2:], counter)

			readyid <- hex.EncodeToString(raw)
		}
	}()
}
