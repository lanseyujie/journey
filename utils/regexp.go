package utils

import "regexp"

const (
    Name   = `^[\w]{8,32}$`
    Email  = `^(?i)[a-z0-9._%+-]+@(?:[a-z0-9-]+\.)+[a-z]{2,6}$`
    Url    = `^(?i)http[s]?://([\w-]+\.)+[\w-]+(:[0-9]{1,4})?(/[\w-./?%&=]*)?$`
    Phone  = `^1[3-9](\d{9})$`
    Number = `^[\d]+$`
    Hash   = `^[a-zA-Z0-9]+$`
)

func IsName(source string) bool {
    return regexp.MustCompile(Name).MatchString(source)
}

func IsEmail(source string) bool {
    return regexp.MustCompile(Email).MatchString(source)
}

func IsUrl(source string) bool {
    return regexp.MustCompile(Url).MatchString(source)
}

func IsPhone(source string) bool {
    return regexp.MustCompile(Phone).MatchString(source)
}

func IsNumber(source string) bool {
    return regexp.MustCompile(Number).MatchString(source)
}

func IsHash(source string) bool {
    return regexp.MustCompile(Hash).MatchString(source)
}
