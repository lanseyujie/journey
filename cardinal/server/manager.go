package server

import (
    "context"
    "flag"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
)

type Manager struct {
    timeout   time.Duration
    service   []Service
    errorChan chan error
}

type Service interface {
    Handler(errorChan chan<- error)
    Release(ctx context.Context)
}

var (
    flagDaemon bool
    daemon     bool
    status     bool
    stop       bool
    restart    bool
    manager    *Manager
)

func init() {
    flagDaemon = os.Getenv("CARDINAL_FLAG_DAEMON") == "true"
    flag.BoolVar(&daemon, "d", false, "daemon")
    flag.BoolVar(&status, "status", false, "status")
    flag.BoolVar(&stop, "stop", false, "stop")
    flag.BoolVar(&restart, "reload", false, "restart")
}

// NewManager
func NewManager() *Manager {
    if manager != nil {
        return manager
    }

    if !flag.Parsed() {
        flag.Parse()
    }

    manager = &Manager{
        timeout:   5 * time.Second,
        errorChan: make(chan error),
    }

    return manager
}

// SetTimeOut
func (m *Manager) SetTimeOut(d time.Duration) *Manager {
    m.timeout = d

    return m
}

// AddService
func (m *Manager) AddService(service ...Service) *Manager {
    m.service = append(m.service, service...)

    return m
}

// Master
func (m *Manager) Master() {
    pid, exist := m.process()
    if daemon {
        if exist {
            log.Println("already running, pid:", pid)
        } else if flagDaemon {
            _ = syscall.Chdir("/")
            syscall.Umask(0)
            m.Worker()
        } else {
            // TODO:// log file
            err := m.daemon("error.log")
            if err != nil {
                log.Println("daemon error:", err)
                os.Exit(-1)
            }
        }
    } else if status {
        if exist {
            log.Println("already running, pid:", pid)
        } else {
            log.Println("process is not running")
        }
    } else if stop {
        if exist {
            err := syscall.Kill(pid, syscall.SIGTERM)
            if err != nil {
                log.Println("syscall.Kill error:", err)
                os.Exit(-1)
            }
        } else {
            log.Println("process is not running")
        }
    } else if restart {
        if exist {
            err := syscall.Kill(pid, syscall.SIGTERM)
            if err != nil {
                log.Println("syscall.Kill error:", err)
                os.Exit(-1)
            }

            // TODO:// log file
            err = m.daemon("error.log")
            if err != nil {
                log.Println("daemon error:", err)
                os.Exit(-1)
            }
        } else {
            log.Println("process is not running")
        }
    } else {
        m.Worker()
    }
}

// Worker
func (m *Manager) Worker() {
    go m.handleSignal()

    for _, s := range m.service {
        go s.Handler(m.errorChan)
    }

    select {
    case err := <-m.errorChan:
        log.Println("exited, error:", err)
        os.Exit(0)
    }
}

// shutdown
func (m *Manager) shutdown() {
    ctx, _ := context.WithTimeout(context.Background(), m.timeout)
    for _, s := range m.service {
        go s.Release(ctx)
    }

    select {
    case <-ctx.Done():
        m.errorChan <- nil
    }
}

// handleSignal
func (m *Manager) handleSignal() {
    ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

    for {
        sig := <-ch
        switch sig {
        case syscall.SIGINT, syscall.SIGTERM:
            log.Println("Received SIGINT or SIGTERM. exiting.")
            m.shutdown()
        default:
            log.Println("Received", sig, ": ignored.")
        }
    }
}

func (m *Manager) process() (int, bool) {
    // TODO:// pid file
    return 0, false
}

func (m *Manager) daemon(filename string) (err error) {
    if os.Getppid() == 1 {
        err = os.Setenv("CARDINAL_FLAG_DAEMON", "true")
    }

    var stdin, stdout, stderr, nullFile, logFile *os.File
    nullFile, err = os.Open(os.DevNull)
    if err != nil {
        return
    }

    if filename != "" {
        logFile, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
        if err != nil {
            return
        }
        stdin = nullFile
        stdout = logFile
        stderr = logFile
    } else {
        stdin = nullFile
        stdout = nullFile
        stderr = nullFile
    }

    dir, _ := os.Getwd()
    procAttr := &syscall.ProcAttr{
        Dir:   dir,
        Env:   os.Environ(),
        Files: []uintptr{stdin.Fd(), stdout.Fd(), stderr.Fd()},
        Sys: &syscall.SysProcAttr{
            Setsid: true,
        },
    }

    // var pid int
    _, err = syscall.ForkExec(os.Args[0], os.Args, procAttr)
    // hide process name
    // pid, err = syscall.ForkExec(os.Args[0], []string{""}, procAttr)
    if err != nil {
        return
    }

    // TODO:// save(pid)
    return
}
