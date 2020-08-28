package template

import (
    "html/template"
    "strings"
    "time"
)

var funcMap = make(template.FuncMap)

func init() {
    AddFuncMap("html", Html)
    AddFuncMap("string", String)
    AddFuncMap("stringjoin", StringJoin)
    AddFuncMap("dateformat", DateFormat)
    AddFuncMap("substr", Substr)
    AddFuncMap("add", Add)
    AddFuncMap("sub", Subtract)
    AddFuncMap("mul", Multiply)
    AddFuncMap("div", Divide)
}

// AddFuncMap register a func in the template
func AddFuncMap(key string, fn interface{}) {
    funcMap[key] = fn
}

// Html
func Html(str string) template.HTML {
    return template.HTML(str)
}

// String
func String(bytes []byte) string {
    return string(bytes)
}

// StringJoin
func StringJoin(strs []string, sep string) string {
    return strings.Join(strs, sep)
}

// DateFormat
func DateFormat(t time.Time, layout string) string {
    return t.Format(layout)
}

// Substr
func Substr(s string, start, length int) string {
    chars := []rune(s)
    if start < 0 {
        start = 0
    }

    if start > len(chars) {
        start = start % len(chars)
    }

    var end int
    if (start + length) > (len(chars) - 1) {
        end = len(chars)
    } else {
        end = start + length
    }

    return string(chars[start:end])
}

// Add
func Add(nums ...int) (ret int) {
    for _, num := range nums {
        ret += num
    }

    return
}

// Subtract
func Subtract(nums ...int) (ret int) {
    for index, num := range nums {
        if index == 0 {
            ret = num
        } else {
            ret -= num
        }
    }

    return
}

// Multiply
func Multiply(nums ...int) (ret int) {
    for index, num := range nums {
        if index == 0 {
            ret = num
        } else {
            ret *= num
        }
    }

    return
}

// Divide
func Divide(nums ...int) (ret int) {
    for index, num := range nums {
        if index == 0 {
            ret = num
        } else {
            ret /= num
        }
    }

    return
}
