package utils

import "strings"

// PascalCase member_id => Member_Id
// the format is camel case if lf[0] is true
// UnderScoreCase: hello_world
// PascalCase: HelloWorld
// CamelCase: helloWorld
func PascalCase(underscore string, lf ...bool) string {
    usc := []byte(strings.ToLower(underscore))
    cc := make([]byte, 0, len(usc))
    lowerFirst := len(lf) == 1 && lf[0]
    flag := false
    for i, ascii := range usc {
        if ascii == '_' {
            flag = true
        } else if ('a' <= ascii && ascii <= 'z') && ((i == 0 && !lowerFirst) || flag) {
            flag = false
            // convert to upper case
            // ASCII A~Z => 65~90 a~z => 97~122
            cc = append(cc, byte(int(ascii)-(97-65)))
        } else {
            cc = append(cc, ascii)
        }
    }

    return string(cc)
}

// UnderScoreCase MemberId => member_id
func UnderScoreCase(camel string) string {
    cc := []byte(camel)
    usc := make([]byte, 0, len(cc))
    for index, ascii := range cc {
        if 'A' <= ascii && ascii <= 'Z' {
            if index > 0 {
                usc = append(usc, '_')
            }
            // convert to lower case
            // ASCII A~Z => 65~90 a~z => 97~122
            usc = append(usc, byte(int(ascii)+97-65))
        } else {
            usc = append(usc, ascii)
        }
    }

    return string(usc)
}
