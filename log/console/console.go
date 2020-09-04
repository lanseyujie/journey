package console

import (
    "fmt"
    "os"
    "sync"
)

type Console struct {
    lock *sync.Mutex
}

// var console = &Console{}

func NewConsole() *Console {
    return &Console{
        lock: &sync.Mutex{},
    }
}

// StringPurple
func StringPurple(str string) string {
    return fmt.Sprintf("%c[%d;%d;%dm%s%c[0m", 0x1B, 0, 35, 1, str, 0x1B)
}

// StringBlue
func StringBlue(str string) string {
    return fmt.Sprintf("%c[%d;%d;%dm%s%c[0m", 0x1B, 0, 34, 1, str, 0x1B)
}

// StringYellow
func StringYellow(str string) string {
    return fmt.Sprintf("%c[%d;%d;%dm%s%c[0m", 0x1B, 0, 33, 1, str, 0x1B)
}

// StringGreen
func StringGreen(str string) string {
    return fmt.Sprintf("%c[%d;%d;%dm%s%c[0m", 0x1B, 0, 32, 1, str, 0x1B)
}

// StringRed
func StringRed(str string) string {
    return fmt.Sprintf("%c[%d;%d;%dm%s%c[0m", 0x1B, 0, 31, 1, str, 0x1B)
}

// Write
func (c *Console) Write(p []byte) (n int, err error) {
    c.lock.Lock()
    defer c.lock.Unlock()

    n = len(p)
    msg := string(p)
    length := len("[INFO]")
    if len(msg) >= length {
        switch msg[:length] {
        case "[INFO]":
            _, err = fmt.Fprint(os.Stdout, StringBlue(msg))
        case "[WARN]":
            _, err = fmt.Fprint(os.Stderr, StringYellow(msg))
        case "[ERRO]", "[FATA]":
            _, err = fmt.Fprint(os.Stderr, StringRed(msg))
        case "[HTTP]":
            _, err = fmt.Fprint(os.Stdout, StringGreen(msg))
        case "[DBUG]":
            _, err = fmt.Fprint(os.Stdout, StringPurple(msg))
        default:
            _, err = fmt.Fprint(os.Stdout, msg)
        }
    } else {
        _, err = fmt.Fprint(os.Stdout, msg)
    }

    return
}
