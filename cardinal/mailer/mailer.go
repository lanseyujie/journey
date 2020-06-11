package mailer

import (
    "crypto/tls"
    "errors"
    "io"
    "net"
    "net/smtp"
    "strconv"
)

type Mailer struct {
    Host     string
    Port     int
    UserName string
    Password string
    Security string // NONE / TLS / STARTTLS
}

// NewMailer
func NewMailer(host string, port int, username, password, security string) *Mailer {
    return &Mailer{
        Host:     host,
        Port:     port,
        UserName: username,
        Password: password,
        Security: security,
    }
}

// NewConnect
func (mailer *Mailer) NewConnect() (conn net.Conn, err error) {
    socket := mailer.Host + ":" + strconv.FormatInt(int64(mailer.Port), 10)
    if mailer.Security == "TLS" {
        conn, err = tls.Dial("tcp", socket, &tls.Config{
            InsecureSkipVerify: true,
            ServerName:         mailer.Host,
        })
    } else {
        conn, err = net.Dial("tcp", socket)
    }

    return
}

// NewClient
func (mailer *Mailer) NewClient(conn net.Conn) (*smtp.Client, error) {
    client, err := smtp.NewClient(conn, mailer.Host)
    if err != nil {
        return nil, err
    }

    err = client.Hello("localhost")
    if err != nil {
        return nil, err
    }

    // StartTLS
    if mailer.Security == "STARTTLS" {
        if ok, _ := client.Extension("STARTTLS"); ok {
            err = client.StartTLS(&tls.Config{
                InsecureSkipVerify: true,
                ServerName:         mailer.Host,
            })
            if err != nil {
                return nil, err
            }
        }
    }

    // Plain Auth
    auth := smtp.PlainAuth("", mailer.UserName, mailer.Password, mailer.Host)
    if ok, _ := client.Extension("AUTH"); ok {
        if err = client.Auth(auth); err != nil {
            return nil, err
        }
    } else {
        return nil, errors.New("smtp: server doesn't support AUTH")
    }

    return client, nil
}

// SendMail
func (mailer *Mailer) SendMail(mail *Mail) (err error) {
    msg := mail.Bytes()

    var conn net.Conn
    conn, err = mailer.NewConnect()
    if err != nil {
        return
    }
    defer conn.Close()

    var client *smtp.Client
    client, err = mailer.NewClient(conn)
    if err != nil {
        return
    }
    defer client.Close()

    sender := mail.From
    if mail.Sender != nil {
        sender = mail.Sender
    }
    err = client.Mail(sender.Address)
    if err != nil {
        return
    }

    // TO
    for _, addr := range mail.To {
        err = client.Rcpt(addr.Address)
        if err != nil {
            return
        }
    }

    // CC
    for _, addr := range mail.Cc {
        err = client.Rcpt(addr.Address)
        if err != nil {
            return
        }
    }

    // BCC
    for _, addr := range mail.Bcc {
        err = client.Rcpt(addr.Address)
        if err != nil {
            return
        }
    }

    // send data
    var w io.WriteCloser
    w, err = client.Data()
    if err != nil {
        return
    }
    _, err = w.Write(msg)
    if err != nil {
        return
    }
    _ = w.Close()

    return client.Quit()
}
