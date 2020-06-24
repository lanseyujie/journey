package log

import (
    "fmt"
    "io"
    "journey/cardinal/log/console"
    "time"
)

type Log struct {
    writer  io.Writer
    disable bool
}

var std = Adapter(console.NewConsole())

// Adapter
func Adapter(writer ...io.Writer) *Log {
    return &Log{
        writer: io.MultiWriter(writer...),
    }
}

// Disable
func (log *Log) Disable() *Log {
    log.disable = true

    return log
}

// Enable
func (log *Log) Enable() *Log {
    log.disable = false

    return log
}

// Write
func (log *Log) Write(v ...interface{}) (n int, err error) {
    if log.writer != nil {
        n, err = fmt.Fprintf(log.writer, "%s %s", time.Now().Format("2006/01/02 15:04:05.000"), fmt.Sprint(v...))
    }

    return
}

// WriteWithPrefix
func (log *Log) WriteWithPrefix(prefix string, v ...interface{}) (n int, err error) {
    if log.writer != nil {
        n, err = fmt.Fprintf(log.writer, "%s %s %s", prefix, time.Now().Format("2006/01/02 15:04:05.000"), fmt.Sprintln(v...))
    }

    return
}

// Println
func (log *Log) Println(v ...interface{}) {
    if !log.disable {
        _, _ = log.Write(fmt.Sprintln(v...))
    }
}

// Info
func (log *Log) Info(v ...interface{}) {
    if !log.disable {
        _, _ = log.WriteWithPrefix("[INFO]", v...)
    }
}

// Warn
func (log *Log) Warn(v ...interface{}) {
    if !log.disable {
        _, _ = log.WriteWithPrefix("[WARN]", v...)
    }
}

// Error
func (log *Log) Error(v ...interface{}) {
    if !log.disable {
        _, _ = log.WriteWithPrefix("[ERRO]", v...)
    }
}

// Http
func (log *Log) Http(v ...interface{}) {
    if !log.disable {
        _, _ = log.WriteWithPrefix("[HTTP]", v...)
    }
}

// SetDefaultLog
func SetDefaultLog(l *Log) {
    std = l
}

// Disable
func Disable() {
    std.disable = true
}

// Enable
func Enable() {
    std.disable = false
}

// Println
func Println(v ...interface{}) {
    if !std.disable {
        _, _ = std.Write(fmt.Sprintln(v...))
    }
}

// Info
func Info(v ...interface{}) {
    if !std.disable {
        _, _ = std.WriteWithPrefix("[INFO]", v...)
    }
}

// Warn
func Warn(v ...interface{}) {
    if !std.disable {
        _, _ = std.WriteWithPrefix("[WARN]", v...)
    }
}

// Error
func Error(v ...interface{}) {
    if !std.disable {
        _, _ = std.WriteWithPrefix("[ERRO]", v...)
    }
}

// Http
func Http(v ...interface{}) {
    if !std.disable {
        _, _ = std.WriteWithPrefix("[HTTP]", v...)
    }
}
