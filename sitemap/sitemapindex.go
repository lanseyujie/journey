package sitemap

import (
    "bytes"
    "encoding/xml"
    "io"
    "time"
)

type Index struct {
    XMLName xml.Name `xml:"sitemapindex"`
    Xmlns   string   `xml:"xmlns,attr"`
    Url     []*Url   `xml:"sitemap"`
}

// NewIndex
func NewIndex() *Index {
    return &Index{
        Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
    }
}

// NewUrl
func (index *Index) NewUrl(url *Url) {
    if url.LastMod.IsZero() {
        url.LastMod = time.Now()
    }
    url.ChangeFreq = ""
    url.Priority = 0
    index.Url = append(index.Url, url)
}

// WriteTo
func (index *Index) WriteTo(w io.Writer) (n int64, err error) {
    var (
        buffer bytes.Buffer
        length int
    )
    _, err = buffer.Write([]byte(xml.Header))
    if err != nil {
        return
    }

    err = xml.NewEncoder(&buffer).Encode(index)
    if err != nil {
        return
    }

    length, err = w.Write(buffer.Bytes())

    return int64(length), err
}

// ReadFrom
func (index *Index) ReadFrom(r io.Reader) (n int64, err error) {
    dec := xml.NewDecoder(r)
    err = dec.Decode(index)
    n = dec.InputOffset()

    return
}
