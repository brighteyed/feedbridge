package kometakolomna

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

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
			Title:       "СШОР Комета: документы",
			Link:        &feeds.Link{Href: "https://kometakolomna.ru/docs/"},
			Description: "СШОР по конькобежному спорту Комета. Протоколы соревнований и присвоение спортивных разрядов",
		},
	}
}

func (p *plugin) Info() pm.PluginMetadata {
	return pm.PluginMetadata{
		TechnicalName: "kometakolomna",
		Name:          "СШОР Комета: документы",
		Description:   "СШОР по конькобежному спорту Комета. Протоколы соревнований и присвоение спортивных разрядов",
		Author:        "brighteyed",
		AuthorURL:     "https://github.com/brighteyed",
		SourceURL:     "https://kometakolomna.ru/docs/",
	}
}

func (p *plugin) Run() (*feeds.Feed, error) {
	var urls = []string{
		"https://kometakolomna.ru/docs/31/",
		"https://kometakolomna.ru/docs/34/",
	}

	result, err := scrape.URLToDocument(p.c, scrape.URLtoTask(urls))
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

		// Documents
		var itemId string
		var docs []string

		for {
			current := s.NextFiltered(".docs-table")

			href := current.Find(".files .tool a")
			path, exists := href.Attr("download")
			if !exists {
				break
			}

			docUrl, err := url.JoinPath("https://kometakolomna.ru", path)
			if err != nil {
				return
			}

			docs = append(docs, fmt.Sprintf(`<a href="%s">Скачать</a>`, docUrl))
			itemId = path
			s = current
		}

		if len(docs) > 0 {
			item.Link = &feeds.Link{Href: "https://kometakolomna.ru/docs/"}
			item.Description = strings.Join(docs, "<br/>")
			item.Id = itemId

			feedItems = append(feedItems, item)
		}
	})
	return feedItems, nil
}
