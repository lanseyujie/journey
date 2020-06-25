package utils

import (
    "bytes"
    "testing"
)

func TestGetAscii(t *testing.T) {
    ascii := GetAscii()
    for i, b := range ascii {
        if byte(i) != b {
            t.Errorf("bytes[%d] = %v, want %v", i, b, byte(i))
            return
        }
    }
}

func TestRandomBytes(t *testing.T) {
    b1 := RandomBytes(16)
    b2 := RandomBytes(16)
    if bytes.Equal(b1, b2) {
        t.FailNow()
    }
}
