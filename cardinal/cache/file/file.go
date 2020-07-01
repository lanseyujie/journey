package file

import (
    "bytes"
    "crypto/md5"
    "encoding/gob"
    "encoding/hex"
    "errors"
    "io"
    "io/ioutil"
    "os"
    "path/filepath"
    "time"
)

type File struct {
    path   string
    period time.Duration
}

var (
    ErrInitDirFailed             = errors.New("cache: file: failed to initialize cache path")
    ErrKeyNotExistOrNotPermanent = errors.New("cache: memory: key not exist or not permanent")
    ErrValueTypeNotInt           = errors.New("cache: file: data type is not integer")
)

// NewFile
func NewFile(path string, period time.Duration) *File {
    return &File{
        path:   path,
        period: period,
    }
}

// getFileName
func (f *File) getFileName(key string) (string, error) {
    h := md5.New()
    if _, err := io.WriteString(h, key); err != nil {
        return "", err
    }

    md5sum := hex.EncodeToString(h.Sum(nil))

    return filepath.Join(f.path, md5sum+".bin"), nil
}

// getCache
func (f *File) getCache(name string) (*Cache, error) {
    bs, err := ioutil.ReadFile(name)
    if err != nil {
        return nil, err
    }

    var cache *Cache
    buffer := bytes.NewBuffer(bs)
    decoder := gob.NewDecoder(buffer)
    err = decoder.Decode(&cache)
    if err != nil {
        return nil, err
    }

    return cache, nil
}

// Init
func (f *File) Init() error {
    if _, err := os.Stat(f.path); err != nil {
        if os.IsNotExist(err) {
            err = nil
            err = os.MkdirAll(f.path, os.ModePerm)
        }
        if err != nil {
            return ErrInitDirFailed
        }
    }

    // disable GC
    if f.period < time.Second*10 {
        return nil
    }

    // GC
    go func() {
        for {
            <-time.After(f.period)

            _ = filepath.Walk(f.path, func(path string, info os.FileInfo, err error) error {
                if info == nil {
                    return nil
                }

                // ignore directories and symbolic links
                if info.IsDir() || (info.Mode()&os.ModeSymlink) > 0 {
                    return nil
                }

                length := len(path)
                if length == 32+4 && path[length-4:] == ".bin" {
                    cache, err := f.getCache(path)
                    if err == nil && cache.Expire() {
                        _ = os.Remove(path)
                    }
                }

                return err
            })
        }
    }()

    return nil
}

// Exist
func (f *File) Exist(key string) bool {
    name, err := f.getFileName(key)
    if err != nil {
        return false
    }

    cache, err := f.getCache(name)
    if err != nil {
        return false
    }

    if cache.Expire() {
        return false
    }

    return true
}

// Get
func (f *File) Get(key string) interface{} {
    name, err := f.getFileName(key)
    if err != nil {
        return nil
    }

    cache, err := f.getCache(name)
    if err != nil {
        return nil
    }

    if cache.Expire() {
        return nil
    }

    return cache.Data
}

// Put
func (f *File) Put(key string, value interface{}, lifetime time.Duration) error {
    gob.Register(value)

    cache := &Cache{
        Data:     value,
        Create:   time.Now(),
        Lifetime: lifetime,
    }
    name, err := f.getFileName(key)
    if err != nil {
        return err
    }

    buffer := bytes.NewBuffer(nil)
    encoder := gob.NewEncoder(buffer)
    err = encoder.Encode(cache)
    if err != nil {
        return err
    }

    err = ioutil.WriteFile(name, buffer.Bytes(), os.ModePerm)
    if err != nil {
        return err
    }

    return nil
}

// Del
func (f *File) Del(key string) error {
    name, err := f.getFileName(key)
    if err != nil {
        return err
    }
    if _, err := os.Stat(name); err != nil {
        if !os.IsNotExist(err) {
            return nil
        }

        return err
    }

    return os.Remove(name)
}

// Incr
func (f *File) Incr(key string) error {
    name, err := f.getFileName(key)
    if err != nil {
        return err
    }

    cache, err := f.getCache(name)
    if os.IsNotExist(err) || (err == nil && cache.Lifetime > 0) {
        return ErrKeyNotExistOrNotPermanent
    }
    if err != nil {
        return err
    }

    switch value := cache.Data.(type) {
    case int:
        cache.Data = value + 1
    case int32:
        cache.Data = value + 1
    case int64:
        cache.Data = value + 1
    case uint:
        cache.Data = value + 1
    case uint32:
        cache.Data = value + 1
    case uint64:
        cache.Data = value + 1
    default:
        return ErrValueTypeNotInt
    }

    return f.Put(key, cache.Data, 0)
}

// Decr
func (f *File) Decr(key string) error {
    name, err := f.getFileName(key)
    if err != nil {
        return err
    }

    cache, err := f.getCache(name)
    if os.IsNotExist(err) || (err == nil && cache.Lifetime > 0) {
        return ErrKeyNotExistOrNotPermanent
    }
    if err != nil {
        return err
    }

    switch value := cache.Data.(type) {
    case int:
        cache.Data = value - 1
    case int32:
        cache.Data = value - 1
    case int64:
        cache.Data = value - 1
    case uint:
        if value > 0 {
            cache.Data = value - 1
        }
    case uint32:
        if value > 0 {
            cache.Data = value - 1
        }
    case uint64:
        if value > 0 {
            cache.Data = value - 1
        }
    default:
        return ErrValueTypeNotInt
    }

    return f.Put(key, cache.Data, 0)
}

// Drop
func (f *File) Drop() (err error) {
    err = filepath.Walk(f.path, func(path string, info os.FileInfo, err error) error {
        if info == nil {
            return err
        }

        // ignore directories and symbolic links
        if info.IsDir() || (info.Mode()&os.ModeSymlink) > 0 {
            return nil
        }

        length := len(path)
        if length == 32+4 && path[length-4:] == ".bin" {
            err = os.Remove(path)
            if err != nil {
                return err
            }
        }

        return err
    })

    return
}
