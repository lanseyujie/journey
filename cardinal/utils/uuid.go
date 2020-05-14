package utils

import (
    "crypto/rand"
    "crypto/sha1"
    "encoding/hex"
    "io"
)

type Uuid [16]byte

// NewV4 returns random generated UUID
func NewUuidV4() (*Uuid, error) {
    u := &Uuid{}
    _, err := io.ReadFull(rand.Reader, u[:])
    if err != nil {
        return nil, err
    }

    var version byte = 4

    // set version bits
    u[6] = (u[6] & 0x0f) | (version << 4)

    // set variant RFC4122 bits
    u[8] = (u[8] & (0xff >> 2)) | (0x02 << 6)

    return u, nil
}

// NewV5 returns UUID based on SHA-1 hash of namespace UUID and name
func NewUuidV5(namespace []byte, name string) (*Uuid, error) {
    u := &Uuid{}
    h := sha1.New()
    _, err := h.Write(namespace)
    if err != nil {
        return nil, err
    }
    _, err = h.Write([]byte(name))
    if err != nil {
        return nil, err
    }

    copy(u[:], h.Sum(nil))

    var version byte = 5

    // set version bits
    u[6] = (u[6] & 0x0f) | (version << 4)

    // set variant RFC4122 bits
    u[8] = (u[8] & (0xff >> 2)) | (0x02 << 6)

    return u, nil
}

// Returns canonical string representation of UUID
// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
func (u *Uuid) String() string {
    buf := make([]byte, 36)

    hex.Encode(buf[0:8], u[0:4])
    buf[8] = '-'
    hex.Encode(buf[9:13], u[4:6])
    buf[13] = '-'
    hex.Encode(buf[14:18], u[6:8])
    buf[18] = '-'
    hex.Encode(buf[19:23], u[8:10])
    buf[23] = '-'
    hex.Encode(buf[24:], u[10:])

    return string(buf)
}
