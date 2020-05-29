package server

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
)

const (
    FlagDaemon   = "CARDINAL_FLAG_DAEMON"
    FlagGraceful = "CARDINAL_FLAG_GRACEFUL"
)

type Manager struct {
    timeout   time.Duration
    logFile   string
    pidFile   *PidFile
    service   []Service
    errorChan chan error
}

type Service interface {
    Handler(errorChan chan<- error)
    Release(ctx context.Context)
}

var (
    flagDaemon     bool
    flagGraceful   bool
    flagRestart    bool
    commandSupport = []string{"-d", "status", "stop", "restart", "reload"}
    command        string
    manager        *Manager
)

func init() {
    flagDaemon = os.Getenv(FlagDaemon) == "true"
    flagGraceful = os.Getenv(FlagGraceful) == "true"

    for _, arg := range os.Args[1:] {
        for _, cmd := range commandSupport {
            if arg == cmd {
                command = cmd
                break
            }
        }
    }
}

// NewManager
func NewManager() *Manager {
    if manager != nil {
        return manager
    }

    manager = &Manager{
        timeout:   2 * time.Second,
        errorChan: make(chan error),
    }

    return manager
}

// SetTimeOut
func (m *Manager) SetTimeOut(d time.Duration) *Manager {
    m.timeout = d

    return m
}

// LogFile
func (m *Manager) LogFile(name string) *Manager {
    m.logFile = name

    return m
}

// AddService
func (m *Manager) AddService(service ...Service) *Manager {
    m.service = append(m.service, service...)

    return m
}

// Master
func (m *Manager) Master() {
    isDaemon := os.Getppid() == 1
    pid, err := m.process()
    if err != nil {
        log.Println("daemon: ", err)

        return
    }

    if command == "-d" {
        if pid > 0 {
            log.Println("daemon: already running, pid is", pid)
        } else if flagDaemon {
            // comment for restart issue here
            // _ = syscall.Chdir("/")
            syscall.Umask(0)
            m.Worker()
        } else {
            err = m.daemon()
            if err != nil {
                log.Println("daemon:", err)
            }
        }
    } else if !isDaemon {
        if command == "status" {
            if pid > 0 {
                log.Println("daemon: already running, pid is", pid)
            } else {
                log.Println("daemon: process is not running")
            }
        } else if command == "stop" {
            if pid > 0 {
                err := syscall.Kill(pid, syscall.SIGTERM)
                if err != nil {
                    log.Println("daemon: syscall.Kill", err)
                }
            } else {
                log.Println("daemon: process is not running")
            }
        } else if command == "restart" {
            if pid > 0 {
                err = syscall.Kill(pid, syscall.SIGUSR1)
                if err != nil {
                    log.Println("daemon: syscall.Kill", err)
                }
            } else {
                log.Println("daemon: process is not running")
            }
        } else {
            if pid > 0 {
                log.Println("daemon: already running, pid is", pid)
            } else {
                m.Worker()
            }
        }
    }

    log.Println("daemon: exited, pid:", os.Getpid())
}

// Worker
func (m *Manager) Worker() {
    go m.handleSignal()

    for _, s := range m.service {
        go s.Handler(m.errorChan)
    }

    select {
    case err := <-m.errorChan:
        if m.pidFile != nil {
            _ = m.pidFile.Release()
        }
        if err != nil {
            log.Println("daemon: error:", err)
        }
        if flagRestart {
            err = m.daemon()
            if err != nil {
                log.Println("daemon: restart,", err)
            }
        }
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
    signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2)
    for {
        sig := <-ch
        switch sig {
        case syscall.SIGINT, syscall.SIGTERM:
            log.Println("Received SIGINT or SIGTERM. exiting.")
            m.shutdown()
        case syscall.SIGHUP:
            log.Println("Received SIGHUP. reloading.")
            // m.graceful()
        case syscall.SIGUSR1:
            log.Println("Received SIGUSR1. restarting.")
            m.restart()
        default:
            log.Println("Received", sig, ": ignored.")
        }
    }
}

// process
func (m *Manager) process() (pid int, err error) {
    m.pidFile, err = NewPidFile("/tmp/cardinal.pid")
    if err != nil {
        return
    }

    err = m.pidFile.Lock()
    if err != nil {
        // already running
        if err == ErrResourceUnavailable {
            return m.pidFile.Read()
        }

        return
    }

    err = m.pidFile.Write()

    return
}

// stdFile
func (m *Manager) stdFile() (stdin, stdout, stderr *os.File, err error) {
    var nullFile, logFile *os.File
    nullFile, err = os.Open(os.DevNull)
    if err != nil {
        return
    }

    if m.logFile != "" {
        logFile, err = os.OpenFile(m.logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
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

    return
}

// daemon
func (m *Manager) daemon() (err error) {
    if os.Getppid() == 1 {
        err = os.Setenv(FlagDaemon, "true")
        if err != nil {
            return
        }
    }

    var stdin, stdout, stderr *os.File
    stdin, stdout, stderr, err = m.stdFile()
    if err != nil {
        return
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

    _, err = syscall.ForkExec(os.Args[0], os.Args, procAttr)

    return
}

// restart
func (m *Manager) restart() {
    flagRestart = true

    m.shutdown()
}
