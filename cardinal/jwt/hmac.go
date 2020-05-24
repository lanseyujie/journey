package jwt

import (
    "crypto"
    "crypto/hmac"
    _ "crypto/sha256" // imports SHA-256 hash function
    _ "crypto/sha512" // imports SHA-384 and SHA-512 hash functions
    "encoding/base64"
    "errors"
    "hash"
)

const (
    // SHA256 is the SHA-256 hashing function
    SHA256 = crypto.Hash(crypto.SHA256)
    // SHA384 is the SHA-384 hashing function
    SHA384 = crypto.Hash(crypto.SHA384)
    // SHA512 is the SHA-512 hashing function
    SHA512 = crypto.Hash(crypto.SHA512)
)

var pool = make([]*HMAC, 0, 1)

// HMAC
type HMAC struct {
    secret string
    crypto crypto.Hash
    hash   hash.Hash
    size   int
}

// NewHMAC creates a new HMAC signing method
func NewHMAC(c crypto.Hash, secret string) *HMAC {
    var hm *HMAC
    if len(pool) > 0 {
        for _, hm = range pool {
            if hm.crypto == c && hm.secret == secret {
                return hm
            }
        }
    }

    h := hmac.New(c.New, []byte(secret))
    hm = &HMAC{
        secret: secret,
        crypto: c,
        hash:   h,
        size:   h.Size(),
    }

    pool = append(pool, hm)

    return hm
}

// String returns the signing method name
func (h *HMAC) String() string {
    switch h.crypto {
    case crypto.SHA256:
        return "HS256"
    case crypto.SHA384:
        return "HS384"
    case crypto.SHA512:
        return "HS512"
    default:
        return ""
    }
}

// Sign signs a hp and returns the signature
func (h *HMAC) Sign(hp []byte) ([]byte, error) {
    if _, err := h.hash.Write(hp); err != nil {
        return nil, err
    }

    return h.hash.Sum(nil), nil
}

// Verify signature
func (h *HMAC) Verify(hp, sign []byte) (err error) {
    var s1, s2 []byte
    b64 := base64.RawURLEncoding
    s1 = make([]byte, b64.DecodedLen(len(sign)))
    _, err = b64.Decode(s1, sign)
    if err != nil {
        return
    }

    s2, err = h.Sign(hp)
    if err != nil {
        return
    }

    if !hmac.Equal(s1, s2) {
        return errors.New("jwt: invalid signature")
    }

    return nil
}
