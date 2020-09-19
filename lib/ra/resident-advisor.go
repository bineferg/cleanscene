package ra

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cleanscene.flights/lib/airports"
	"github.com/cleanscene.flights/lib/crawler"
	"github.com/cleanscene.flights/lib/event"
)

type RA interface {
	LoadArtists(string) (map[string]Artist, error)
	LoadEvents(Artist) (map[time.Time]event.Event, error)
}

func New(airSvc airports.Airports, crwlr crawler.Crawler, outputDir, tourYear string) RA {
	return ResidentAdvisor{
		airSvc:    airSvc,
		crawler:   crwlr,
		outputDir: outputDir,
		tourYear:  tourYear,
	}
}

type residentAdvisor struct {
	airSvc    airports.Airports
	crawler   crawler.Crawler
	tourYear  string
	outputDir string
}

func (ra residentAdvisor) LoadArtists(fileName string) (map[string]Artist, error) {
	var artists = make(map[string]Artist)

	file, err := os.Open(fileName)
	if err != nil {
		return artists, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		arr := strings.Split(scanner.Text(), ",")
		eCount, _ := strconv.Atoi(arr[3])
		link, _ := ra.crawler.GetArtistUrl(arr[0])
		airCode, _ := ra.airSvc.AirCodeByCity(arr[1], arr[2])
		artists[arr[0]] = Artist{
			Name:        arr[0],
			City:        arr[1],
			Country:     arr[2],
			EventsTotal: eCount,
			Link:        link,
			AirCode:     airCode,
			// initialise with empty events map
			Events: make(map[time.Time]event.Event),
		}

	}

	if err := scanner.Err(); err != nil {
		return artists, err
	}
	return artists, nil

}

type Artist struct {
	Name        string
	City        string
	Country     string
	EventsTotal int
	Link        string
	AirCode     string
	Events      Events
}

type Events map[time.Time]event.Event

func (ra residentAdvisor) getEventAirports(events Events) Events {
	for d, e := range events {
		edge, err := ra.airSvc.FindClosestAirport(e.Location)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		e.AirCode = edge.Code
		e.City = edge.CityCode
		e.Country = edge.Country
		events[d] = e
	}
	return events
}

func (ra residentAdvisor) LoadEvents(a Artist) (Events, error) {
	events, err := ra.crawler.GetEvents(a.Link, ra.tourYear)
	if err != nil {
		return events, err
	}
	return ra.getEventAirports(events), nil
}
