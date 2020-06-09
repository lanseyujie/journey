package mailer

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

func NewMail() *Mail {
    return &Mail{}
}
