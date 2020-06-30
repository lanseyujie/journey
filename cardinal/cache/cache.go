package cache

import "time"

type Cache interface {
    Init() error
    Exist(key string) bool
    Get(key string) interface{}
    Put(key string, value interface{}, lifetime time.Duration) error
    Del(key string) error
    Incr(key string) error
    Decr(key string) error
    Drop() error
}

var adapters = make(map[string]Cache)

// Register
func Register(name string, cache Cache) {
    if cache == nil {
        panic("cache: adapter is nil")
    }
    if _, ok := adapters[name]; ok {
        panic("cache: register adapter repeatedly")
    }
    adapters[name] = cache
}

// Get
func Get(name string) (cache Cache, exist bool) {
    cache, exist = adapters[name]

    return
}
