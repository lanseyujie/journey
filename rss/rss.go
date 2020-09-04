package rss

import (
    "bytes"
    "encoding/xml"
    "io"
    "time"
)

type Rss struct {
    XMLName xml.Name `xml:"rss"`
    Xmlns   string   `xml:"xmlns:dc,attr"`
    Version string   `xml:"version,attr"`
    Channel *Channel `xml:"channel"`
}

type Channel struct {
    Title       string  `xml:"title"`
    Link        string  `xml:"link"`
    Description string  `xml:"description"`
    PubDate     string  `xml:"pubDate"`
    Generator   string  `xml:"generator,omitempty"`
    Item        []*Item `xml:"item"`
}

type Item struct {
    Title       string `xml:"title"`
    Link        string `xml:"link"`
    Category    string `xml:"category,omitempty"`
    Description string `xml:"description"`
    PubDate     string `xml:"pubDate"`
    Guid        string `xml:"guid,omitempty"`
}

// NewRss
func NewRss() *Rss {
    return &Rss{
        Xmlns:   "http://purl.org/dc/elements/1.1/",
        Version: "2.0",
        Channel: &Channel{
            PubDate:   time.Now().Format(time.RFC822Z),
            Generator: "Cardinal RSS Generator 1.0",
        },
    }
}

// NewChannel
func (rss *Rss) NewChannel(title, link, desc string) {
    rss.Channel.Title = title
    rss.Channel.Link = link
    rss.Channel.Description = desc
}

// NewItem
func (rss *Rss) NewItem(item *Item) {
    if item.PubDate == "" {
        item.PubDate = time.Now().Format(time.RFC822Z)
    }
    rss.Channel.Item = append(rss.Channel.Item, item)
}

// WriteTo
func (rss *Rss) WriteTo(w io.Writer) (n int64, err error) {
    var (
        buffer bytes.Buffer
        length int
    )
    _, err = buffer.Write([]byte(xml.Header))
    if err != nil {
        return
    }

    err = xml.NewEncoder(&buffer).Encode(rss)
    if err != nil {
        return
    }

    length, err = w.Write(buffer.Bytes())

    return int64(length), err
}

// ReadFrom
func (rss *Rss) ReadFrom(r io.Reader) (n int64, err error) {
    dec := xml.NewDecoder(r)
    err = dec.Decode(rss)
    n = dec.InputOffset()

    return
}
