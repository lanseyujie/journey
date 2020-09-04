package jwt

import (
    "crypto"
    "encoding/base64"
    "encoding/json"
    "errors"
    "strings"
    "time"
)

/*
jwt token
header.payload.signature
base64(headerJson).base64(payloadJson).base64(hash(base64(headerJson).base64(payloadJson), secret))
*/

type Header struct {
    Algorithm string `json:"alg,omitempty"` // encryption algorithm
    Type      string `json:"typ,omitempty"` // JWT
}

type Payload struct {
    Issuer     string      `json:"iss,omitempty"`
    Subject    string      `json:"sub,omitempty"`
    Audience   []string    `json:"aud,omitempty"`
    Expiration int64       `json:"exp,omitempty"`
    NotBefore  int64       `json:"nbf,omitempty"`
    IssuedAt   int64       `json:"iat,omitempty"`
    JwtId      string      `json:"jti,omitempty"`
    Data       interface{} `json:"dat,omitempty"` // custom data
}

type Raw struct {
    Header    string
    Payload   string
    Signature string
}

type Jwt struct {
    Header    *Header
    Payload   *Payload
    Signature []byte
    Raw       *Raw
    checked   bool
}

func NewJwt() *Jwt {
    return &Jwt{
        Header:    &Header{},
        Payload:   &Payload{},
        Signature: []byte{},
        Raw:       &Raw{},
    }
}

// Sign the JWT token
func (jwt *Jwt) Sign(hmac *HMAC) (token []byte, err error) {
    var hJson, pJson []byte

    jwt.Header.Algorithm = hmac.String()
    jwt.Header.Type = "JWT"

    hJson, err = json.Marshal(jwt.Header)
    if err != nil {
        return
    }
    pJson, err = json.Marshal(jwt.Payload)
    if err != nil {
        return
    }

    // url safe base64 encoding
    b64 := base64.RawURLEncoding
    // get the length of each part
    h64len := b64.EncodedLen(len(hJson))
    p64len := b64.EncodedLen(len(pJson))
    s64len := b64.EncodedLen(hmac.size)

    token = make([]byte, h64len+1+p64len+1+s64len)
    // base64 encode header json
    b64.Encode(token, hJson)
    token[h64len] = '.'
    // base64 encode payload json
    b64.Encode(token[h64len+1:], pJson)
    // hmac sign base64(headerJson).base64(payloadJson)
    jwt.Signature, err = hmac.Sign(token[:h64len+1+p64len])
    if err != nil {
        return nil, err
    }
    token[h64len+1+p64len] = '.'
    // base64 encode signature
    b64.Encode(token[h64len+1+p64len+1:], jwt.Signature)

    return
}

// Verify signature
func (jwt *Jwt) Verify(token string, secret string) (err error) {
    var h64Json, p64Json []byte
    raw := strings.Split(token, ".")
    if len(raw) != 3 {
        return errors.New("jwt: malformed token")
    }

    jwt.Raw.Header = raw[0]
    jwt.Raw.Payload = raw[1]
    jwt.Raw.Signature = raw[2]

    // url safe base64 encoding
    b64 := base64.RawURLEncoding

    // decode header
    h64Json, err = b64.DecodeString(jwt.Raw.Header)
    if err != nil {
        return
    }
    err = json.Unmarshal(h64Json, jwt.Header)
    if err != nil {
        return
    }

    // decode payload
    p64Json, err = b64.DecodeString(jwt.Raw.Payload)
    if err != nil {
        return
    }
    err = json.Unmarshal(p64Json, jwt.Payload)
    if err != nil {
        return
    }

    // decode signature
    jwt.Signature, err = b64.DecodeString(jwt.Raw.Signature)
    if err != nil {
        return
    }

    var hash crypto.Hash
    switch jwt.Header.Algorithm {
    case "HS256":
        hash = SHA256
    case "HS384":
        hash = SHA384
    case "HS512":
        hash = SHA512
    default:
        return errors.New("jwt: unsupported hash algorithm")
    }

    hmac := NewHMAC(hash, secret)
    err = hmac.Verify([]byte(jwt.Raw.Header+"."+jwt.Raw.Payload), []byte(jwt.Raw.Signature))
    if err != nil {
        return
    }

    jwt.checked = true

    return
}

// CheckIssuer
func (jwt *Jwt) CheckIssuer(iss string) bool {
    if jwt.checked && jwt.Payload.Issuer == iss {
        return true
    }

    return false
}

// CheckSubject
func (jwt *Jwt) CheckSubject(sub string) bool {
    if jwt.checked && jwt.Payload.Subject == sub {
        return true
    }

    return false
}

// CheckAudience
// It checks if at least one of the audiences in the JWT's payload is listed in aud
func (jwt *Jwt) CheckAudience(aud []string) bool {
    if jwt.checked {
        for _, serverAud := range aud {
            for _, clientAud := range jwt.Payload.Audience {
                if clientAud == serverAud {
                    return true
                }
            }
        }
    }

    return false
}

// CheckExpiration
func (jwt *Jwt) CheckExpiration() bool {
    if jwt.checked {
        if jwt.Payload.Expiration == 0 || time.Now().Before(time.Unix(jwt.Payload.Expiration, 0)) {
            return true
        }
    }

    return false
}

// CheckNotBefore
func (jwt *Jwt) CheckNotBefore() bool {
    if jwt.checked {
        if jwt.Payload.NotBefore == 0 || time.Now().After(time.Unix(jwt.Payload.NotBefore, 0)) {
            return true
        }
    }

    return false
}

// CheckIssuedAt
func (jwt *Jwt) CheckIssuedAt() bool {
    if jwt.checked {
        if jwt.Payload.IssuedAt == 0 || time.Now().After(time.Unix(jwt.Payload.IssuedAt, 0)) {
            return true
        }
    }

    return false
}

// CheckJwtId
func (jwt *Jwt) CheckJwtId(jti string) bool {
    if jwt.checked && jwt.Payload.JwtId == jti {
        return true
    }

    return false
}

// GetCustomData
func (jwt *Jwt) GetCustomData() interface{} {
    if jwt.checked {
        return jwt.Payload.Data
    }

    return nil
}
