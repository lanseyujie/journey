package server

import (
    "errors"
    "fmt"
    "io"
    "os"
    "syscall"
)

var ErrResourceUnavailable = errors.New("daemon: resource temporarily unavailable")

type PidFile struct {
    *os.File
}

// NewPidFile
func NewPidFile(name string) (*PidFile, error) {
    file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0640)
    if err != nil {
        return nil, err
    }

    return &PidFile{File: file}, nil
}

// Lock the pid file
func (pf *PidFile) Lock() (err error) {
    err = syscall.Flock(int(pf.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
    if err == syscall.EWOULDBLOCK {
        err = ErrResourceUnavailable
    }

    return
}

// Unlock the pid file
func (pf *PidFile) Unlock() (err error) {
    err = syscall.Flock(int(pf.Fd()), syscall.LOCK_UN)
    if err == syscall.EWOULDBLOCK {
        err = ErrResourceUnavailable
    }

    return
}

// Set writes current pid to file
func (pf *PidFile) Set() (err error) {
    err = pf.Lock()
    if err != nil {
        return
    }

    _, err = pf.Seek(0, io.SeekStart)
    if err != nil {
        return
    }

    var fileLen int
    fileLen, err = fmt.Fprint(pf.File, os.Getpid())
    if err != nil {
        return
    }

    err = pf.Truncate(int64(fileLen))
    if err != nil {
        return
    }

    return pf.Sync()
}

// Get the running pid
func (pf *PidFile) Get() (pid int, err error) {
    err = pf.Lock()
    if err != nil {
        // already running
        if err == ErrResourceUnavailable {
            _, err = pf.Seek(0, io.SeekStart)
            if err != nil {
                return
            }
            _, err = fmt.Fscan(pf.File, &pid)
            if err == io.EOF {
                err = nil
            }
        }
    }

    return
}

// Release the pid file
func (pf *PidFile) Release() (err error) {
    err = pf.Unlock()
    if err != nil {
        return
    }

    _ = pf.Close()

    return os.Remove(pf.Name())
}
