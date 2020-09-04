package sitemap

import (
    "bytes"
    "testing"
    "time"
)

func TestSitemap_WriteTo(t *testing.T) {
    smap := NewSitemap()
    smap.NewUrl(&Url{
        Loc:        "http://example.com/post/1.html",
        LastMod:    time.Date(2020, 7, 25, 11, 26, 03, 0, time.Local),
        ChangeFreq: Weekly,
        Priority:   0.2,
    })
    smap.NewUrl(&Url{
        Loc: "http://example.com/post/2.html",
    })

    var buffer bytes.Buffer
    _, err := smap.WriteTo(&buffer)
    if err != nil {
        t.Fatal(err)
    }
}

func TestSitemap_ReadFrom(t *testing.T) {
    smap := NewSitemap()
    xml := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
    <url>
        <loc>http://example.com/post/1.html</loc>
        <lastmod>2020-07-25T11:26:03+08:00</lastmod>
        <changefreq>weekly</changefreq>
        <priority>0.2</priority>
    </url>
    <url>
        <loc>http://example.com/post/2.html</loc>
        <lastmod>2020-07-25T11:33:33.237997555+08:00</lastmod>
    </url>
</urlset>`

    buffer := bytes.NewBuffer([]byte(xml))
    _, err := smap.ReadFrom(buffer)
    if err != nil {
        t.Fatal(err)
    }

    if len(smap.Url) != 2 {
        t.Fatal("decode failed")
    }
}
