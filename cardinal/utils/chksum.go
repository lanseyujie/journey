package utils

import (
    "crypto/md5"
    "crypto/sha1"
    "crypto/sha256"
    "encoding/hex"
    "errors"
    "hash"
    "io"
    "os"
)

var ErrChkSumUnsupportedType = errors.New("chksum: unsupported type")

func Md5sum(b interface{}) (string, error) {
    return ChkSum(md5.New(), b)
}

func Sha1sum(b interface{}) (string, error) {
    return ChkSum(sha1.New(), b)
}

func Sha256sum(b interface{}) (string, error) {
    return ChkSum(sha256.New(), b)
}

// support []byte, string, *os.File
func ChkSum(h hash.Hash, b interface{}) (sum string, err error) {
    switch v := b.(type) {
    case []byte:
        _, err = h.Write(v)
    case string:
        _, err = io.WriteString(h, v)
    case *os.File:
        _, err = io.Copy(h, v)
    default:
        return "", ErrChkSumUnsupportedType
    }

    if err != nil {
        return "", err
    }

    return hex.EncodeToString(h.Sum(nil)), nil
}
