package crawler

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cleanscene.flights/lib/event"
)

type Crawler interface {
	GetArtistUrl(string) (string, error)
	GetEvents(string, string) (map[time.Time]event.Event, error)
}

func New(baseUrl string) (Crawler, error) {
	resp, err := http.Get(baseUrl + "/dj.aspx")
	if err != nil {
		return DJCrawler{}, err
	}
	defer resp.Body.Close()
	var artistLinkMap = make(map[string]string)

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
				artistLinkMap[artistName] = baseUrl + href

			}
		})
	}
	return djCrawler{
		baseUrl:     baseUrl,
		artistLinks: artistLinkMap,
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
	document.Find("a").Each(func(i int, s *goquery.Selection) {
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
	})
	return location, nil
}

func parseEventText(es string) (time.Time, string) {

	// Extract and parse date
	date, i := event.ParseDate(es, 0)

	// Find where the rest of the info is
	title, _ := event.ParseTitle(es, i)
	return date, title
}

func (c djCrawler) GetEvents(artistUrl, tourYear string) (map[time.Time]event.Event, error) {
	events := make(map[time.Time]event.Event)
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
		document.Find("article").Each(func(i int, s *goquery.Selection) {
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
		})
	}
	return events, err
}
