package sitemap

import (
    "bytes"
    "encoding/xml"
    "io"
    "time"
)

type Sitemap struct {
    XMLName xml.Name `xml:"urlset"`
    Xmlns   string   `xml:"xmlns,attr"`
    Url     []*Url   `xml:"url"`
}

type Url struct {
    Loc        string     `xml:"loc"`
    LastMod    time.Time  `xml:"lastmod,omitempty"`
    ChangeFreq ChangeFreq `xml:"changefreq,omitempty"`
    Priority   float32    `xml:"priority,omitempty"`
}

type ChangeFreq string

const (
    Always  ChangeFreq = "always"
    Hourly  ChangeFreq = "hourly"
    Daily   ChangeFreq = "daily"
    Weekly  ChangeFreq = "weekly"
    Monthly ChangeFreq = "monthly"
    Yearly  ChangeFreq = "yearly"
    Never   ChangeFreq = "never"
)

// NewSitemap
func NewSitemap() *Sitemap {
    return &Sitemap{
        Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
    }
}

// NewUrl
func (smap *Sitemap) NewUrl(url *Url) {
    if url.LastMod.IsZero() {
        url.LastMod = time.Now()
    }
    smap.Url = append(smap.Url, url)
}

// WriteTo
func (smap *Sitemap) WriteTo(w io.Writer) (n int64, err error) {
    var (
        buffer bytes.Buffer
        length int
    )
    _, err = buffer.Write([]byte(xml.Header))
    if err != nil {
        return
    }

    err = xml.NewEncoder(&buffer).Encode(smap)
    if err != nil {
        return
    }

    length, err = w.Write(buffer.Bytes())

    return int64(length), err
}

// ReadFrom
func (smap *Sitemap) ReadFrom(r io.Reader) (n int64, err error) {
    dec := xml.NewDecoder(r)
    err = dec.Decode(smap)
    n = dec.InputOffset()

    return
}
