package swling

import (
	"net/http"
	"strings"

	pm "github.com/dewey/feedbridge/plugin"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
	"golang.org/x/net/html"
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
		l: log.With(l, "plugin", "swling"),
		c: c,
		f: &feeds.Feed{
			Title:       "swling.ru",
			Link:        &feeds.Link{Href: "https://swling.ru"},
			Description: "Поговорим о радио?",
		},
	}
}

func (p *plugin) Info() pm.PluginMetadata {
	return pm.PluginMetadata{
		TechnicalName: "swling",
		Name:          p.f.Title,
		Description:   "Радиопанорама. Журнал о событиях в мире радио и других средств связи",
		Author:        "brighteyed",
		AuthorURL:     "https://github.com/brighteyed",
		SourceURL:     "https://swling.ru",
	}
}

func (p *plugin) Run() (*feeds.Feed, error) {
	req, err := http.NewRequest(http.MethodGet, "https://swling.ru/feed", nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.c.Do(req)
	if err != nil {
		return nil, err
	}

	fp := gofeed.NewParser()
	feed, err := fp.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var feedItems []*feeds.Item
	for _, fi := range feed.Items {
		audio, err := extractAudioFile(fi.Content)
		if err != nil {
			return nil, err
		}
		if audio == "" {
			continue
		}

		audioResp, err := http.Get(audio)
		if err != nil {
			return nil, err
		}

		item := &feeds.Item{
			Author: &feeds.Author{
				Name:  fi.Author.Name,
				Email: fi.Author.Email,
			},
			Title: fi.Title,
			Link: &feeds.Link{
				Href: fi.Link,
			},
			Id:          fi.GUID,
			Description: prepareDescription(fi.Content),
			Enclosure: &feeds.Enclosure{
				Url:    audio,
				Length: audioResp.Header.Get("Content-Length"),
				Type:   "audio/mpeg",
			},
		}
		if fi.PublishedParsed != nil {
			item.Created = *fi.PublishedParsed
		}
		if fi.UpdatedParsed != nil {
			item.Updated = *fi.UpdatedParsed
		}
		feedItems = append(feedItems, item)
	}
	p.f.Items = feedItems
	return p.f, nil
}

func extractAudioFile(s string) (string, error) {
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return "", err
	}

	var audioLink string

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "audio" {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "a" {
					for _, a := range c.Attr {
						if a.Key == "href" {
							audioLink = a.Val
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return audioLink, nil
}

func prepareDescription(s string) string {
	parts := strings.Split(s, "</ul>")
	return strings.Join(parts[:len(parts)-1], "</ul>") + "</ul>"
}
