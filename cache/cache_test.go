package cache

import (
    "github.com/lanseyujie/journey/cache/file"
    "github.com/lanseyujie/journey/cache/memory"
    "testing"
    "time"
)

func TestMemoryCache(t *testing.T) {
    mem := memory.NewMemory(time.Second * 10)
    Register("memory", mem)

    err := mem.Init()
    if err != nil {
        t.Fatal(err)
    }

    cache, exist := Get("memory")
    if !exist {
        t.Fatal(exist)
    }

    num := 100
    if err = cache.Put("num", num, time.Second*3); err != nil {
        t.Fatal(err)
    }
    if n := cache.Get("num"); n != num {
        t.Fatalf("cache.Get() = %v, want %v", n, num)
    }
    time.Sleep(time.Second * 3)
    if n := cache.Get("num"); n != nil {
        t.Fatalf("cache.Get() = %v, want %v", n, nil)
    }

    name := "jike"
    if err = cache.Put("name", name, 0); err != nil {
        t.Fatal(err)
    }
    if n := cache.Get("name"); n != name {
        t.Fatalf("cache.Get() = %v, want %v", n, name)
    }
    if err = cache.Del("name"); err != nil {
        t.Fatal(err)
    }
    if cache.Exist("name") {
        t.Fatal("cache exist")
    }

    if err = cache.Put("num", num, 0); err != nil {
        t.Fatal(err)
    }
    if err = cache.Decr("num"); err != nil {
        t.Fatal(err)
    }
    if n := cache.Get("num"); n != 99 {
        t.Fatalf("cache.Get() = %v, want %v", n, 99)
    }

    if err = cache.Drop(); err != nil {
        t.Fatal(err)
    }
}

func TestFileCache(t *testing.T) {
    f := file.NewFile("./cache", time.Second*10)
    Register("file", f)

    err := f.Init()
    if err != nil {
        t.Fatal(err)
    }

    cache, exist := Get("file")
    if !exist {
        t.Fatal(exist)
    }

    num := 100
    if err = cache.Put("num", num, time.Second*3); err != nil {
        t.Fatal(err)
    }
    if n := cache.Get("num"); n != num {
        t.Fatalf("cache.Get() = %v, want %v", n, num)
    }
    time.Sleep(time.Second * 3)
    if n := cache.Get("num"); n != nil {
        t.Fatalf("cache.Get() = %v, want %v", n, nil)
    }

    name := "jike"
    if err = cache.Put("name", name, 0); err != nil {
        t.Fatal(err)
    }
    if n := cache.Get("name"); n != name {
        t.Fatalf("cache.Get() = %v, want %v", n, name)
    }
    if err = cache.Del("name"); err != nil {
        t.Fatal(err)
    }
    if cache.Exist("name") {
        t.Fatal("cache exist")
    }

    if err = cache.Put("num", num, 0); err != nil {
        t.Fatal(err)
    }
    if err = cache.Decr("num"); err != nil {
        t.Fatal(err)
    }
    if n := cache.Get("num"); n != 99 {
        t.Fatalf("cache.Get() = %v, want %v", n, 99)
    }

    if err = cache.Drop(); err != nil {
        t.Fatal(err)
    }
}
