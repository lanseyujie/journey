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
func ChkSum(h hash.Hash, b interface{}) (string, error) {
    var bytes []byte

    if f, ok := b.(*os.File); ok {
        if _, err := io.Copy(h, f); err != nil {
            return "", err
        }
    } else {
        if bs, ok := b.([]byte); ok {
            bytes = bs
        } else if str, ok := b.(string); ok {
            bytes = []byte(str)
        } else {
            return "", errors.New("chksum: unsupported type")
        }

        if _, err := h.Write(bytes); err != nil {
            return "", err
        }
    }

    return hex.EncodeToString(h.Sum(nil)), nil
}
