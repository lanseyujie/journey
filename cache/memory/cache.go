package memory

import "time"

type Cache struct {
    data     interface{}
    create   time.Time
    lifetime time.Duration
}

// Expire
func (c *Cache) Expire() bool {
    if c.lifetime == 0 {
        return false
    }

    return time.Now().Sub(c.create) > c.lifetime
}
