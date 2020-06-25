package utils

import (
    "crypto/rand"
    rnd "math/rand"
    "time"
)

var alphabet = []byte(`0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz`)

// RandomBytes generate random []byte
func RandomBytes(n int, chars ...byte) []byte {
    var bytes = make([]byte, n)
    var flag bool

    length := len(chars)
    if length == 0 {
        chars = alphabet
        length = len(chars)
    }
    maxIndex := length - 1

    if num, err := rand.Read(bytes); num != n || err != nil {
        rnd.Seed(time.Now().UnixNano())
        flag = true
    }

    for i, b := range bytes {
        if flag {
            bytes[i] = chars[rnd.Intn(maxIndex)]
        } else {
            bytes[i] = chars[b%byte(maxIndex)]
        }
    }

    return bytes
}
