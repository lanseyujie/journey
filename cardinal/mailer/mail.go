package mailer

import (
    "mime"
    "regexp"
)

type Addressee struct {
    Name    string
    Address string
}

type Body struct {
    ContentType string
    Content     string
}

type Attachment struct {
    Name        string
    Type        string
    ContentId   string
    Description string
    MimeType    string
    FilePath    string
}

type Mail struct {
    From         Addressee
    ReplyTo      Addressee
    To           []Addressee
    Cc           []Addressee
    Bcc          []Addressee
    Subject      string
    Organization string
    Priority     int
    MessageId    string
    Body         Body
    Attachment   []Attachment
    Boundary     string
}

var IsAddress = regexp.MustCompile(`^(?i)[a-z0-9._%+-]+@(?:[a-z0-9-]+\.)+[a-z]{2,6}$`)

func NewMail() *Mail {
    return &Mail{}
}

// String
func (addr *Addressee) String() string {
    if IsAddress.MatchString(addr.Address) {
        return mime.BEncoding.Encode("utf-8", addr.Name) + " <" + addr.Address + ">"
    }
    // return "\"" + mime.BEncoding.Encode("utf-8", addr.Name) + "\" <" + addr.Address + ">"

    return ""
}
