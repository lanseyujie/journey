package cache

import "time"

type Cache interface {
    Exist(key string) bool
    Get(key string) interface{}
    Put(key string, value interface{}, timeout time.Duration) error
    Del(key string) error
    Incr(key string) error
    Decr(key string) error
    Drop() error
}
