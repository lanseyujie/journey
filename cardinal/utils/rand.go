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

    if len(chars) == 0 {
        chars = alphabet
    }

    if num, err := rand.Read(bytes); num != n || err != nil {
        rnd.Seed(time.Now().UnixNano())
        flag = true
    }

    for i, b := range bytes {
        if flag {
            bytes[i] = chars[rnd.Intn(len(chars))]
        } else {
            bytes[i] = chars[b%byte(len(chars))]
        }
    }

    return bytes
}
