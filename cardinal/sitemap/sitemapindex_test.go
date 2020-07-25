package sitemap

import (
    "bytes"
    "testing"
)

func TestIndex_WriteTo(t *testing.T) {
    index := NewIndex()
    index.NewUrl(&Url{
        Loc: "http://example.com/sitemap-1.xml",
    })
    index.NewUrl(&Url{
        Loc: "http://example.com/sitemap-2.xml",
    })

    var buffer bytes.Buffer
    _, err := index.WriteTo(&buffer)
    if err != nil {
        t.Fatal(err)
    }
}

func TestIndex_ReadFrom(t *testing.T) {
    index := NewIndex()
    xml := `<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
    <sitemap>
        <loc>http://example.com/sitemap-1.xml</loc>
        <lastmod>2020-07-25T14:40:08.83792185+08:00</lastmod>
    </sitemap>
    <sitemap>
        <loc>http://example.com/sitemap-2.xml</loc>
        <lastmod>2020-07-25T14:40:08.837922821+08:00</lastmod>
    </sitemap>
</sitemapindex>`

    buffer := bytes.NewBuffer([]byte(xml))
    _, err := index.ReadFrom(buffer)
    if err != nil {
        t.Fatal(err)
    }

    if len(index.Url) != 2 {
        t.Fatal("decode failed")
    }
}
