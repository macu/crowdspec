package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"

	crypto_rand "crypto/rand"
)

func randomToken(length uint) string {
	randGen := rand.New(rand.NewSource(randomSeed()))

	bytes := make([]byte, length)

	for i := range bytes {
		bytes[i] = sessionRandLetters[randGen.Intn(len(sessionRandLetters))]
	}

	return string(bytes)
}

func randomSeed() int64 {
	var randomBytes [8]byte
	_, err := crypto_rand.Read(randomBytes[:])

	if err == nil {

		// Hopefully this is always the code path in production
		// since session IDs could be guessed by nano times

		var seed int64
		err = binary.Read(bytes.NewBuffer(randomBytes[:]), binary.LittleEndian, &seed)
		if err == nil {
			return seed
		}

		logError(nil, nil, fmt.Errorf("reading session ID random generator seed with binary.Read: %w", err))
		return time.Now().UnixNano()

	}

	logError(nil, nil, fmt.Errorf("reading session ID random generator seed bytes with crypto/rand: %w", err))
	return time.Now().UnixNano()
}
