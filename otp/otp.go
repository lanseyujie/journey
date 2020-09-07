package otp

import (
    "crypto/hmac"
    "crypto/sha1"
    "encoding/base32"
    "encoding/binary"
    "github.com/lanseyujie/journey/utils"
    "net/url"
    "strconv"
    "strings"
    "time"
)

type Otp struct {
    counter int64  // htop counter
    scope   int64  // tolerance scope, times or minutes
    secret  string // secret key
}

// NewOtp
func NewOtp(scope int64, secret string, begin ...int64) *Otp {
    var cnt int64 = -1
    if len(begin) == 1 && begin[0] >= 0 {
        cnt = begin[0]
    }

    return &Otp{
        counter: cnt,
        scope:   scope,
        secret:  secret,
    }
}

// ComputeCode
func (otp *Otp) ComputeCode(secret string, ts int64) (code int64, err error) {
    var bytes []byte
    bytes, err = base32.StdEncoding.DecodeString(strings.TrimSpace(secret))
    if err != nil {
        return
    }

    buf := make([]byte, 8)
    binary.BigEndian.PutUint64(buf, uint64(ts))
    // buf, err = hex.DecodeString(fmt.Sprintf("%016x", ts))
    // if err != nil {
    //     return
    // }

    h := hmac.New(sha1.New, bytes)
    _, err = h.Write(buf)
    if err != nil {
        return
    }

    sum := h.Sum(nil)

    // https://tools.ietf.org/html/rfc4226#section-5.4
    offset := sum[len(sum)-1] & 0x0f
    truncated := binary.BigEndian.Uint32(sum[offset : offset+4])
    truncated &= 0x7fffffff
    code = int64(truncated % 1000000)

    return
}

// VerifyTotpCode
func (otp *Otp) VerifyTotpCode(code int64) bool {
    ts := time.Now().Unix()
    min := (ts / 30) - otp.scope
    max := (ts / 30) + otp.scope

    for t := min; t <= max; t++ {
        if num, err := otp.ComputeCode(otp.secret, t); err == nil && num == code {
            return true
        }
    }

    return false
}

// VerifyHotpCode
func (otp *Otp) VerifyHotpCode(code int64) bool {
    min := otp.counter - otp.scope
    max := otp.counter + otp.scope
    for i := min; i <= max; i++ {
        if num, err := otp.ComputeCode(otp.secret, i); err == nil && num == code {
            otp.counter++

            return true
        }
    }

    otp.counter++

    return false
}

// GetCount returns the HOTP counter
func (otp *Otp) GetCount() int64 {
    return otp.counter
}

// Auth
func (otp *Otp) Auth(code string) bool {
    if len(code) == 6 {
        num, err := strconv.ParseInt(code, 10, 64)
        if err != nil {
            return false
        }

        if otp.counter == -1 {
            return otp.VerifyTotpCode(num)
        } else {
            return otp.VerifyHotpCode(num)
        }
    }

    return false
}

// KeyGen Generate key
func KeyGen() string {
    alphabets := []byte(`0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ`)

    return base32.StdEncoding.EncodeToString(utils.RandomBytes(25, alphabets...))
}

// GetUri for QR code
func (otp *Otp) GetUri(account, issuer string) string {
    q := make(url.Values)
    auth := "totp/"
    if otp.counter >= 0 {
        auth = "hotp/"
        q.Add("counter", strconv.FormatInt(otp.counter, 10))
    }

    q.Add("secret", otp.secret)

    if issuer != "" {
        q.Add("issuer", issuer)
        auth += issuer + ":"
    }

    return "otpauth://" + auth + account + "?" + q.Encode()
}
