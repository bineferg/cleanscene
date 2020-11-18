package crawler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cleanscene.flights/lib/event"
)

type Crawler interface {
	GetArtistUrl(string) (string, error)
	GetArtistEvents(string, string) (Events, error)
}

func New(url string) (Crawler, error) {
	resp, err := http.Get(fmt.Sprintf("%s/dj.aspx", url))
	if err != nil {
		return djCrawler{}, err
	}
	defer resp.Body.Close()
	var artistUrls = make(map[string]string)

	if resp.StatusCode == http.StatusOK {
		document, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal("Error loading HTTP response body. ", err)
		}
		// Find all links
		document.Find("a").Each(func(index int, element *goquery.Selection) {

			// See if the href attribute exists on the element
			href, _ := element.Attr("href")
			if strings.Contains(href, "/dj/") && !strings.Contains(href, "favourites") {
				artistName := element.Text()
				artistUrls[artistName] = url + href

			}
		})
	}
	return djCrawler{
		baseUrl:     url,
		artistLinks: artistUrls,
	}, nil

}

type djCrawler struct {
	baseUrl     string
	artistLinks map[string]string
}

func (c djCrawler) GetArtistUrl(name string) (string, error) {
	link, ok := c.artistLinks[name]
	if !ok {
		return link, errors.New("artist link not found!")
	}
	return link, nil

}

var scrapeClubFunc = func(location string, document *goquery.Document) func(i int, s *goquery.Selection) {
	return func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if s.Text() == "Google Maps" && strings.Contains(href, "maps") {
			location = href
			return
		} else {
			document.Find("span").Each(func(i int, s *goquery.Selection) {
				ip, _ := s.Attr("itemprop")
				if ip == "street-address" {
					location = s.Text()
				}
			})
		}
	}
}

func (c djCrawler) findClubLocation(clubLink string) (string, error) {
	resp, err := http.Get(c.baseUrl + clubLink)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}
	var location string
	document.Find("a").Each(scrapeClubFunc(location, document))
	return location, nil
}

func parseEventText(es string) (time.Time, string) {
	// Extract and parse date
	date, i := event.ParseDate(es, 0)

	// Find where the rest of the info is
	title, _ := event.ParseTitle(es, i)
	return date, title
}

type Events map[time.Time]event.Event

var eventScrapeFunc = func(events Events, c djCrawler) func(i int, s *goquery.Selection) {
	return func(i int, s *goquery.Selection) {
		ip, _ := s.Attr("class")
		if ip == "event" {
			date, title := parseEventText(s.Text())
			s.Find("a").Each(func(i int, element *goquery.Selection) {
				href, _ := element.Attr("href")
				if strings.Contains(href, "/club") {
					location, err := c.findClubLocation(href)
					if err != nil {
						err = err
					}
					events[date] = event.Event{
						Title:    title,
						Location: location,
					}
				}

			})

		}
	}
}

func (c djCrawler) GetArtistEvents(artistUrl, tourYear string) (Events, error) {
	events := make(Events)
	resp, err := http.Get(artistUrl + "/dates?yr=" + tourYear)
	if err != nil {
		return events, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		document, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal("Error loading HTTP response body. ", err)
		}
		document.Find("article").Each(eventScrapeFunc(events, c))

	}
	return events, err
}
