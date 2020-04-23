package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/jakobvarmose/crypta/ipfs"
)

type Entry struct {
	Loc        string
	LastMod    time.Time
	ChangeFreq ChangeFreq
	Priority   float64
}

type ChangeFreq string

const (
	ChangeFreq_Always  ChangeFreq = "always"
	ChangeFreq_Hourly  ChangeFreq = "hourly"
	ChangeFreq_Daily   ChangeFreq = "daily"
	ChangeFreq_Weekly  ChangeFreq = "weekly"
	ChangeFreq_Monthly ChangeFreq = "monthly"
	ChangeFreq_Yearly  ChangeFreq = "yearly"
	ChangeFreq_Never   ChangeFreq = "never"
)

func write(w io.Writer, entries []Entry) error {
	w.Write([]byte(xml.Header))
	e := xml.NewEncoder(w)
	e.EncodeToken(xml.StartElement{xml.Name{"", "urlset"}, []xml.Attr{
		xml.Attr{xml.Name{"", "xmlns"}, "http://www.sitemaps.org/schemas/sitemap/0.9"},
	}})

	for _, entry := range entries {
		e.EncodeToken(xml.CharData("\n  "))
		e.EncodeToken(xml.StartElement{xml.Name{"", "url"}, nil})

		e.EncodeToken(xml.CharData("\n    "))
		e.EncodeToken(xml.StartElement{xml.Name{"", "loc"}, nil})
		e.EncodeToken(xml.CharData(entry.Loc))
		e.EncodeToken(xml.EndElement{xml.Name{"", "loc"}})

		if (entry.LastMod != time.Time{}) {
			e.EncodeToken(xml.CharData("\n    "))
			e.EncodeToken(xml.StartElement{xml.Name{"", "lastmod"}, nil})
			e.EncodeToken(xml.CharData(entry.LastMod.UTC().Format(time.RFC3339)))
			e.EncodeToken(xml.EndElement{xml.Name{"", "lastmod"}})
		}

		if entry.ChangeFreq != "" {
			e.EncodeToken(xml.CharData("\n    "))
			e.EncodeToken(xml.StartElement{xml.Name{"", "changefreq"}, nil})
			e.EncodeToken(xml.CharData(entry.ChangeFreq))
			e.EncodeToken(xml.EndElement{xml.Name{"", "changefreq"}})
		}

		e.EncodeToken(xml.CharData("\n    "))
		e.EncodeToken(xml.StartElement{xml.Name{"", "priority"}, nil})
		e.EncodeToken(xml.CharData(strconv.FormatFloat(entry.Priority, 'f', -1, 64)))
		e.EncodeToken(xml.EndElement{xml.Name{"", "priority"}})

		e.EncodeToken(xml.CharData("\n  "))
		e.EncodeToken(xml.EndElement{xml.Name{"", "url"}})
	}

	e.EncodeToken(xml.CharData("\n"))
	e.EncodeToken(xml.EndElement{xml.Name{"", "urlset"}})
	e.EncodeToken(xml.CharData("\n"))
	return e.Flush()
}

func main() {
	server := "https://crypta.io"
	output := os.Stdout

	repoPath, err := ipfs.BestKnownPath()
	if err != nil {
		fmt.Println(err)
		return
	}

	files, err := ioutil.ReadDir(path.Join(repoPath, "userstore"))
	if err != nil {
		fmt.Println(err)
		return
	}

	var entries []Entry
	entries = append(entries, Entry{
		Loc:      server + "/",
		Priority: 0.7,
	})
	for _, file := range files {
		entries = append(entries, Entry{
			Loc:      server + "/user/" + file.Name(),
			Priority: 0.5,
		})
	}

	write(output, entries)
	return

	//output, err := os.Open("/var/www/html/sitemap.xml")
	//output, err := os.Create("sitemap.xml")
	//if err != nil {
	//	panic(err)
	//}
	//defer output.Close()

	output.Write([]byte(xml.Header))
	e := xml.NewEncoder(output)
	e.EncodeToken(xml.StartElement{xml.Name{"", "urlset"}, []xml.Attr{
		xml.Attr{
			xml.Name{"", "xmlns"},
			"http://www.sitemaps.org/schemas/sitemap/0.9",
		},
	}})

	for _, file := range files {
		e.EncodeToken(xml.CharData("\n  "))
		e.EncodeToken(xml.StartElement{xml.Name{"", "url"}, nil})
		e.EncodeToken(xml.CharData("\n    "))
		e.EncodeToken(xml.StartElement{xml.Name{"", "loc"}, nil})
		e.EncodeToken(xml.CharData(server + "/user/" + file.Name()))
		e.EncodeToken(xml.EndElement{xml.Name{"", "loc"}})
		e.EncodeToken(xml.CharData("\n  "))
		e.EncodeToken(xml.EndElement{xml.Name{"", "url"}})
	}

	e.EncodeToken(xml.CharData("\n"))
	e.EncodeToken(xml.EndElement{xml.Name{"", "urlset"}})
	e.EncodeToken(xml.CharData("\n"))
	e.Flush()
}
