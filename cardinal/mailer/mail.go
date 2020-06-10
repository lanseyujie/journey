package mailer

import (
    "bytes"
    "crypto/md5"
    "encoding/base64"
    "encoding/hex"
    "io/ioutil"
    "mime"
    "os"
    "path/filepath"
    "regexp"
    "strconv"
    "strings"
    "time"
)

const (
    MaxLength = 76
    CRLF      = "\r\n"
)

type Addressee struct {
    Name    string
    Address string
}

type AddresseeList []*Addressee

type Body struct {
    Text string
    Html string
}

type Attachment struct {
    IsInline    bool
    Name        string
    ContentId   string
    Description string
    MimeType    string
    FilePath    string
}

type Mail struct {
    From         *Addressee // author
    Sender       *Addressee // sender
    ReplyTo      *Addressee
    To           AddresseeList
    Cc           AddresseeList // carbon copy
    Bcc          AddresseeList // blind carbon copy
    Subject      string
    Organization string
    Priority     int
    Body         *Body
    Attachment   []*Attachment
    messageId    string
    boundary     string
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

    return ""
}

// String
func (al AddresseeList) String() string {
    var list []string
    for _, value := range al {
        list = append(list, value.String())
    }

    return strings.Join(list, ", ")
}

func base64Wrap(content []byte, buffer *bytes.Buffer) {
    encode := base64.StdEncoding.EncodeToString(content)
    for index, length := 0, len(encode)-1; index <= length; index++ {
        buffer.WriteByte(encode[index])
        // 76 characters per line, excluding \r\n, see RFC 2045
        if (index+1)%MaxLength == 0 {
            buffer.WriteString(CRLF)
        }
    }

    buffer.WriteString(CRLF)
}

func (m *Mail) parseHeader(buffer *bytes.Buffer) {
    header := "From: " + m.From.String() + CRLF
    if m.Sender != nil {
        header += "Sender: " + m.Sender.String() + CRLF
    }
    if m.ReplyTo != nil {
        header += "Reply-To: " + m.ReplyTo.String() + CRLF
    }
    if m.Cc != nil {
        header += "Cc: " + m.Cc.String() + CRLF
    }
    header += "To: " + m.To.String() + CRLF
    if m.Cc != nil {
        header += "Cc: " + m.Cc.String() + CRLF
    }
    // some service providers may expose this field in emails, such as sina mail
    // if m.Bcc != nil {
    //     header += "Bcc: " + m.Bcc.String() + CRLF
    // }

    if m.Subject == "" {
        m.Subject = "No Subject"
    }
    header += "Subject: " + mime.BEncoding.Encode("utf-8", m.Subject) + CRLF
    header += "MessageId: " + m.getMessageId() + CRLF
    header += "Organization: " + mime.BEncoding.Encode("utf-8", m.Organization) + CRLF
    header += "X-Mailer: Cardinal Mailer" + CRLF
    header += "X-Priority: " + strconv.FormatInt(int64(m.Priority), 10) + CRLF
    header += "MIME-Version: 1.0" + CRLF

    isMultiPart := false
    typ := "text/plain" // only text exist
    if (m.Body.Text != "" && m.Body.Html != "") || len(m.Attachment) > 0 {
        isMultiPart = true
        typ = "multipart/alternative" // text and html exist
        if len(m.Attachment) > 0 {
            typ = "multipart/mixed" // attachment exist
            for _, att := range m.Attachment {
                if att.IsInline {
                    typ = "multipart/related" // inline attachment exist
                    break
                }
            }
        }
    } else if m.Body.Html != "" {
        typ = "text/html" // only html exist
    }

    if isMultiPart {
        header += "Content-Type: " + typ + "; Boundary=\"" + m.getBoundary() + "\"; charset=UTF-8" + CRLF
        header += "Content-Transfer-Encoding: 8bit" + CRLF
    } else {
        header += "Content-Type: " + typ + "; charset=UTF-8" + CRLF
        header += "Content-Transfer-Encoding: base64" + CRLF
    }

    header += "Date: " + time.Now().Format(time.RFC1123Z) + CRLF
    header += CRLF

    if isMultiPart {
        header += "This is a multi-part message in MIME format." + CRLF
    }

    buffer.WriteString(header)
}

func (m *Mail) parseBody(buffer *bytes.Buffer) {
    if m.boundary != "" {
        // multipart
        if m.Body.Text != "" {
            header := CRLF
            header += "--" + m.boundary + CRLF
            header += "Content-Type: text/plain; charset=UTF-8" + CRLF
            header += "Content-Transfer-Encoding: base64" + CRLF
            header += CRLF
            buffer.WriteString(header)
            base64Wrap([]byte(m.Body.Text), buffer)
        }
        if m.Body.Html != "" {
            header := CRLF
            header += "--" + m.boundary + CRLF
            header += "Content-Type: text/html; charset=UTF-8" + CRLF
            header += "Content-Transfer-Encoding: base64" + CRLF
            header += CRLF
            buffer.WriteString(header)
            base64Wrap([]byte(m.Body.Html), buffer)
        }
    } else if m.Body.Text != "" {
        base64Wrap([]byte(m.Body.Text), buffer)
    } else if m.Body.Html != "" {
        base64Wrap([]byte(m.Body.Html), buffer)
    }
}

func (m *Mail) parseAttachment(buffer *bytes.Buffer) {
    for _, att := range m.Attachment {
        file, err := ioutil.ReadFile(att.FilePath)
        if err == nil {
            ext := filepath.Ext(att.FilePath)
            if att.MimeType == "" {
                att.MimeType = mime.TypeByExtension(ext)
                if att.MimeType == "" {
                    att.MimeType = "application/octet-stream"
                }
            }

            if att.IsInline && att.Name == "" {
                att.Name = att.ContentId + ext
            }

            header := CRLF
            header += "--" + m.boundary + CRLF
            header += "Content-Type: " + att.MimeType + "; name=\"" + att.Name + "\"" + CRLF
            header += "Content-Transfer-Encoding: base64" + CRLF

            if att.Description != "" {
                header += "Content-Description: " + mime.BEncoding.Encode("utf-8", att.Description) + CRLF
            }

            if att.IsInline {
                header += "Content-Id: <" + att.ContentId + ">" + CRLF
                header += "Content-Disposition: inline; filename=\"" + att.Name + "\"" + CRLF
            } else {
                header += "Content-Disposition: attachment; filename=\"" + att.Name + "\"" + CRLF
            }

            header += CRLF
            buffer.WriteString(header)

            base64Wrap(file, buffer)
        }
    }
}

func (m *Mail) parseFooter(buffer *bytes.Buffer) {
    if m.boundary != "" {
        buffer.WriteString(CRLF + "--" + m.boundary + "--")
    }
}

func (m *Mail) getMessageId() string {
    if m.messageId == "" {
        ts := strconv.FormatInt(time.Now().UnixNano(), 10)
        host, err := os.Hostname()
        if err != nil {
            host = m.Organization
        }

        m.messageId = "<" + ts + ".sender@" + host + ">"
    }

    return m.messageId
}

func (m *Mail) getBoundary() string {
    if m.boundary == "" {
        h := md5.New()
        h.Write([]byte(m.Subject))
        m.boundary = hex.EncodeToString(h.Sum(nil))
    }

    return m.boundary
}

// Bytes
func (m *Mail) Bytes() []byte {
    if m.From == nil || m.To == nil {
        return nil
    }

    buffer := bytes.NewBuffer(make([]byte, 0, 4096))
    m.parseHeader(buffer)
    m.parseBody(buffer)
    m.parseAttachment(buffer)
    m.parseFooter(buffer)

    return buffer.Bytes()
}
