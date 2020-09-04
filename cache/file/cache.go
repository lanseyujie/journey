package file

import "time"

type Cache struct {
    Data     interface{}
    Create   time.Time
    Lifetime time.Duration
}

// Expire
func (c *Cache) Expire() bool {
    if c.Lifetime == 0 {
        return false
    }

    return time.Now().Sub(c.Create) > c.Lifetime
}
