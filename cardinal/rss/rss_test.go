package rss

import (
    "bytes"
    "testing"
    "time"
)

func TestRss_WriteTo(t *testing.T) {
    rss := NewRss()
    rss.NewChannel("EXAMPLE WEB", "https://example.com", "a demo web")
    rss.NewItem(&Item{
        Title:       "hello world",
        Link:        "http://example.com/post/1.html",
        Description: "<b>你好世界</b>",
        PubDate:     time.Now().Format(time.RFC822Z),
    })
    rss.NewItem(&Item{
        Title:       "about",
        Link:        "http://example.com/post/2.html",
        Description: "about this web",
        PubDate:     time.Now().Format(time.RFC822Z),
    })

    var buffer bytes.Buffer
    _, err := rss.WriteTo(&buffer)
    if err != nil {
        t.Fatal(err)
    }
}

func TestRss_ReadFrom(t *testing.T) {
    rss := NewRss()
    xml := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
    <channel>
        <title>EXAMPLE WEB</title>
        <link>https://example.com</link>
        <description>a demo web</description>
        <pubDate>25 Jul 20 15:47 +0800</pubDate>
        <generator>Cardinal RSS Generator 1.0</generator>
        <item>
            <title>hello world</title>
            <link>http://example.com/post/1.html</link>
            <description>&lt;b&gt;你好世界&lt;/b&gt;</description>
            <pubDate>25 Jul 20 15:47 +0800</pubDate>
        </item>
        <item>
            <title>about</title>
            <link>http://example.com/post/2.html</link>
            <description>about this web</description>
            <pubDate>25 Jul 20 15:47 +0800</pubDate>
        </item>
    </channel>
</rss>`

    buffer := bytes.NewBuffer([]byte(xml))
    _, err := rss.ReadFrom(buffer)
    if err != nil {
        t.Fatal(err)
    }

    if len(rss.Channel.Item) != 2 {
        t.Fatal("decode failed")
    }
}
