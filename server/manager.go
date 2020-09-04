package server

import (
    "context"
    "encoding/json"
    "log"
    "net"
    "os"
    "os/signal"
    "syscall"
    "time"
)

const (
    FlagDaemon   = "CARDINAL_FLAG_DAEMON"
    FlagGraceful = "CARDINAL_FLAG_GRACEFUL"
    PidFileName  = "/tmp/cardinal.pid"
)

type Manager struct {
    timeout   time.Duration
    logFile   *os.File
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
    commandSupport []string
    command        string
    addrOrder      []string
    manager        *Manager
)

func init() {
    flagDaemon = os.Getenv(FlagDaemon) == "true"
    flagGraceful = os.Getenv(FlagGraceful) == "true"
    commandSupport = []string{"-d", "status", "stop", "restart", "reload"}
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
func (m *Manager) LogFile(name string) (err error) {
    m.logFile, err = os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

    return
}

// PidFile
func (m *Manager) PidFile(name string) (err error) {
    m.pidFile, err = NewPidFile(name)

    return
}

// AddService
func (m *Manager) AddService(service ...Service) *Manager {
    m.service = append(m.service, service...)

    return m
}

// Master
func (m *Manager) Master() {
    if m.pidFile == nil {
        var err error
        m.pidFile, err = NewPidFile(PidFileName)
        if err != nil {
            log.Println("daemon: NewPidFile,", err)
        }
    }
    pid, err := m.pidFile.Get()
    if err != nil {
        log.Println("daemon: m.pidFile.Get,", err)

        return
    }

    if command == "-d" {
        if pid > 0 {
            if flagGraceful {
                _ = os.Unsetenv(FlagGraceful)
                syscall.Umask(0)
                // get server address order
                decoder := json.NewDecoder(os.Stdin)
                err = decoder.Decode(&addrOrder)
                if err != nil {
                    log.Println("daemon: decoder.Decode error,", err)

                    return
                }

                m.Worker()
            } else {
                log.Println("daemon: already running, pid is", pid)
            }
        } else if flagDaemon {
            _ = os.Unsetenv(FlagDaemon)
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
    } else if os.Getppid() != 1 {
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
        } else if command == "reload" {
            if pid > 0 {
                err = syscall.Kill(pid, syscall.SIGHUP)
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

    log.Println("exited, pid:", os.Getpid())
}

// Worker
func (m *Manager) Worker() {
    // monitor signal
    go m.handleSignal()

    // run service
    for _, s := range m.service {
        go s.Handler(m.errorChan)
    }

    // successful start, update pid to file
    go func() {
        if flagGraceful {
            err := syscall.Kill(os.Getppid(), syscall.SIGTERM)
            if err != nil {
                log.Println("graceful: syscall.Kill error,", err)
                m.errorChan <- err

                return
            }

            // wait for the old process to unlock
            <-time.After(m.timeout + time.Millisecond*200)
        }

        err := m.pidFile.Set()
        if err != nil {
            log.Println("daemon: m.pidRecord error,", err)
            m.errorChan <- err
        }
    }()

    select {
    case err := <-m.errorChan:
        if err != nil {
            log.Println("daemon: exit,", err)
        }

        // do not delete the pid file now to simplify the graceful reload logic
        // _ = m.pidFile.Release()
        _ = m.pidFile.Unlock()

        // restart
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
    // prevent ErrServerClosed error during graceful reload
    if !flagGraceful {
        for _, s := range m.service {
            go s.Release(ctx)
        }
    }

    select {
    case <-ctx.Done():
        m.errorChan <- nil
    }
}

// handleSignal
func (m *Manager) handleSignal() {
    ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGCHLD, syscall.SIGUSR1, syscall.SIGUSR2)
    for {
        sig := <-ch
        switch sig {
        case syscall.SIGINT, syscall.SIGTERM:
            log.Println("Received SIGINT or SIGTERM. exiting.")
            m.shutdown()
        case syscall.SIGHUP:
            log.Println("Received SIGHUP. reloading.")
            m.graceful()
        case syscall.SIGCHLD:
            log.Println("Received SIGCHLD. cleaning.")
        case syscall.SIGUSR1:
            log.Println("Received SIGUSR1. restarting.")
            m.restart()
        default:
            log.Println("Received", sig, ": ignored.")
        }
    }
}

// stdFile
func (m *Manager) stdFile() (stdin, stdout, stderr *os.File, err error) {
    var nullFile *os.File
    nullFile, err = os.Open(os.DevNull)
    if err != nil {
        return
    }

    if m.logFile != nil {
        stdin = nullFile
        stdout = m.logFile
        stderr = m.logFile
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

// graceful
func (m *Manager) graceful() {
    var (
        stdin, stdout, stderr, rPipe, wPipe, socket *os.File
        err                                         error
    )

    err = os.Setenv(FlagGraceful, "true")
    if err != nil {
        log.Println("graceful: os.Setenv error,", err)
        return
    }

    // used to send data to the child process
    rPipe, wPipe, err = os.Pipe()
    if err != nil {
        log.Println("graceful: os.Pipe error,", err)

        return
    }
    defer func() {
        _ = rPipe.Close()
        _ = wPipe.Close()
    }()
    stdin = rPipe

    _, stdout, stderr, err = m.stdFile()
    if err != nil {
        log.Println("graceful:", err)

        return
    }

    // do not close std files, prevent restart command to quit the process due to graceful reload failure

    addrs := make([]string, 0, 2)
    files := []uintptr{stdin.Fd(), stdout.Fd(), stderr.Fd()}
    for _, srv := range cluster {
        ln, ok := srv.listener.(*net.TCPListener)
        if !ok {
            log.Println("listener is not tcp listener")

            return
        }

        // get listener socket
        socket, err = ln.File()
        if err != nil {
            log.Println("get listener socket error:", err)

            return
        }

        addrs = append(addrs, srv.Addr)
        files = append(files, socket.Fd())
    }

    dir, _ := os.Getwd()
    procAttr := &syscall.ProcAttr{
        Dir:   dir,
        Env:   os.Environ(),
        Files: files,
        Sys: &syscall.SysProcAttr{
            Setsid: true,
        },
    }

    var pid int
    pid, err = syscall.ForkExec(os.Args[0], os.Args, procAttr)
    if err != nil {
        log.Println("graceful: syscall.ForkExec,", err)

        return
    }

    go func() {
        proc, err := os.FindProcess(pid)
        if err == nil {
            // if the child process exits abnormally, the child process will be killed,
            // otherwise, the parent process will be killed
            _, _ = proc.Wait()
        }
        _ = syscall.Kill(pid, syscall.SIGKILL)
        _ = os.Unsetenv(FlagGraceful)
    }()

    // send server order list to child process
    encoder := json.NewEncoder(wPipe)
    err = encoder.Encode(addrs)
    if err != nil {
        log.Println("graceful: encoder.Encode error,", err)
    }
}
