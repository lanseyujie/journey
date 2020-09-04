package utils

import (
    "fmt"
    "reflect"
    "runtime"
    "strings"
)

// GetFunctionName
func GetFunctionName(fn interface{}, seps ...rune) string {
    // get function name
    rValue := reflect.ValueOf(fn)
    if rValue.Type().Kind() != reflect.Func {
        return ""
    }

    fullName := runtime.FuncForPC(rValue.Pointer()).Name()
    // use seps to split strings
    fields := strings.FieldsFunc(fullName, func(sep rune) bool {
        for _, s := range seps {
            if sep == s {
                return true
            }
        }

        return false
    })

    if size := len(fields); size > 0 {
        return fields[size-1]
    }

    return "???"
}

// StackTrace
func StackTrace(err error, skip ...int) string {
    pc := make([]uintptr, 16)

    sk := 3
    if len(skip) == 1 && skip[0] >= 0 {
        sk = skip[0]
    }

    n := runtime.Callers(sk, pc)
    frames := runtime.CallersFrames(pc[:n])
    str := strings.Builder{}
    str.WriteString(err.Error() + "\nStackTrace:")
    for {
        frame, more := frames.Next()
        str.WriteString(fmt.Sprintf("\n\t%s:%d %s", frame.File, frame.Line, frame.Function))
        if !more {
            break
        }
    }

    return str.String()
}
