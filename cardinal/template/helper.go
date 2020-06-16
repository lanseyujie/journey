package template

import (
    "html/template"
    "time"
)

// Html
func Html(str string) template.HTML {
    return template.HTML(str)
}

// String
func String(bytes []byte) string {
    return string(bytes)
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
