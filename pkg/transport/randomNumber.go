package transport

import (
	"math/rand"
	"time"
)

func GenerateRandomNumber() uint32 {
	rand.Seed(time.Now().Unix())
	randomNum := rand.Uint32()
	return randomNum
}
