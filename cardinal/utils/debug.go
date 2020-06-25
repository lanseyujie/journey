package utils

import (
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
