package kometakolomna

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	pm "github.com/dewey/feedbridge/plugin"
	"github.com/dewey/feedbridge/scrape"
	"golang.org/x/net/html/charset"

	"github.com/go-kit/kit/log"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
)

// Plugin defines a new plugin
type plugin struct {
	l log.Logger
	c *http.Client
	f *feeds.Feed
}

// NewPlugin initializes a new plugin
func NewPlugin(l log.Logger, c *http.Client) *plugin {
	return &plugin{
		l: log.With(l, "plugin", "kometakolomna"),
		c: c,
		f: &feeds.Feed{
			Title:       "СШОР Комета: протоколы соревнований",
			Link:        &feeds.Link{Href: "https://kometakolomna.ru/docs/31"},
			Description: "СШОР по конькобежному спорту Комета. Протоколы соревнований",
		},
	}
}

func (p *plugin) Info() pm.PluginMetadata {
	return pm.PluginMetadata{
		TechnicalName: "kometakolomna",
		Name:          "СШОР Комета. Протоколы соревнований",
		Description:   `СШОР по конькобежному спорту Комета. Протоколы соревнований`,
		Author:        "brighteyed",
		AuthorURL:     "https://github.com/brighteyed",
		SourceURL:     "https://kometakolomna.ru/docs/31",
	}
}

func (p *plugin) Run() (*feeds.Feed, error) {
	const docsUrl = "https://kometakolomna.ru/docs/31"

	result, err := scrape.URLToDocument(p.c, scrape.URLtoTask([]string{docsUrl}))
	if err != nil {
		return nil, err
	}

	var feedItems []*feeds.Item
	for _, r := range result {
		items, err := p.listHandler(&r.Document, r.ContentType)
		if err != nil {
			p.l.Log("err", err)
		}
		feedItems = append(feedItems, items...)
	}

	p.f.Items = feedItems

	return p.f, nil
}

func (p *plugin) listHandler(doc *goquery.Document, contentType string) ([]*feeds.Item, error) {
	var feedItems []*feeds.Item
	author := &feeds.Author{Name: "feedbridge"}

	doc.Find(".area h3").Each(func(i int, s *goquery.Selection) {
		item := &feeds.Item{
			Author: author,
		}

		// Title
		var header string
		input := bytes.NewReader([]byte(s.Text()))
		output, e := charset.NewReader(input, contentType)
		if e == nil {
			r, _ := ioutil.ReadAll(output)
			header = string(r)
		}
		item.Title = header

		// Document
		t := s.NextFiltered(".docs-table")
		href := t.Find(".files .tool a")
		path, exists := href.Attr("download")
		if !exists {
			return
		}

		docUrl, err := url.JoinPath("https://kometakolomna.ru/docs/31", path)
		if err != nil {
			return
		}

		item.Description = fmt.Sprintf(`<a href="%s">Скачать протоколы</a>`, docUrl)
		item.Link = &feeds.Link{Href: "https://kometakolomna.ru/docs/31"}
		item.Id = path

		feedItems = append(feedItems, item)
	})
	return feedItems, nil
}
