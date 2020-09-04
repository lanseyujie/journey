package memory

import (
    "errors"
    "sync"
    "time"
)

type Memory struct {
    sync.RWMutex
    cache  map[string]*Cache
    period time.Duration
}

var (
    ErrKeyNotExistOrNotPermanent = errors.New("cache: memory: key not exist or not permanent")
    ErrValueTypeNotInt           = errors.New("cache: memory: data type is not integer")
)

// NewMemory
func NewMemory(period time.Duration) *Memory {
    return &Memory{
        cache:  map[string]*Cache{},
        period: period,
    }
}

// Init
func (mem *Memory) Init() error {
    // disable GC
    if mem.period < time.Second*10 {
        return nil
    }

    // GC
    go func() {
        for {
            keys := make([]string, 0, len(mem.cache))
            <-time.After(mem.period)

            mem.RLock()
            for key, cache := range mem.cache {
                if cache.Expire() {
                    keys = append(keys, key)
                }
            }
            mem.RUnlock()

            mem.Lock()
            for _, key := range keys {
                delete(mem.cache, key)
            }
            mem.Unlock()
        }
    }()

    return nil
}

// Exist
func (mem *Memory) Exist(key string) bool {
    mem.RLock()
    defer mem.RUnlock()
    if cache, exist := mem.cache[key]; exist {
        return !cache.Expire()
    }

    return false
}

// Get
func (mem *Memory) Get(key string) interface{} {
    mem.RLock()
    defer mem.RUnlock()

    if cache, exist := mem.cache[key]; exist && !cache.Expire() {
        return cache.data
    }

    return nil
}

// Put
func (mem *Memory) Put(key string, value interface{}, lifetime time.Duration) error {
    mem.Lock()
    mem.cache[key] = &Cache{
        data:     value,
        create:   time.Now(),
        lifetime: lifetime,
    }
    mem.Unlock()

    return nil
}

// Del
func (mem *Memory) Del(key string) error {
    mem.Lock()
    delete(mem.cache, key)
    mem.Unlock()

    return nil
}

// Incr
func (mem *Memory) Incr(key string) error {
    mem.Lock()
    defer mem.Unlock()

    cache, exist := mem.cache[key]
    if !exist || (exist && cache.lifetime > 0) {
        return ErrKeyNotExistOrNotPermanent
    }

    switch value := cache.data.(type) {
    case int:
        cache.data = value + 1
    case int32:
        cache.data = value + 1
    case int64:
        cache.data = value + 1
    case uint:
        cache.data = value + 1
    case uint32:
        cache.data = value + 1
    case uint64:
        cache.data = value + 1
    default:
        return ErrValueTypeNotInt
    }

    return nil
}

// Decr
func (mem *Memory) Decr(key string) error {
    mem.Lock()
    defer mem.Unlock()

    cache, exist := mem.cache[key]
    if !exist || (exist && cache.lifetime > 0) {
        return ErrKeyNotExistOrNotPermanent
    }

    switch value := cache.data.(type) {
    case int:
        cache.data = value - 1
    case int32:
        cache.data = value - 1
    case int64:
        cache.data = value - 1
    case uint:
        if value > 0 {
            cache.data = value - 1
        }
    case uint32:
        if value > 0 {
            cache.data = value - 1
        }
    case uint64:
        if value > 0 {
            cache.data = value - 1
        }
    default:
        return ErrValueTypeNotInt
    }

    return nil
}

// Drop
func (mem *Memory) Drop() error {
    mem.Lock()
    mem.cache = map[string]*Cache{}
    mem.Unlock()

    return nil
}
