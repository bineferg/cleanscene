package flight

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/cleanscene.flights/lib/event"
	"github.com/cleanscene.flights/lib/ra"
	country_mapper "github.com/pirsquare/country-mapper"
)

type Trips []Trip

type Trip struct {
	DepCode string
	ArrCode string
	Date    string
}

type Planner interface {
	Plan(ra.Artist) (Trips, error)
}

func NewPlanner(countryClient *country_mapper.CountryInfoClient) Planner {
	return FlightPlanner{
		cc: countryClient,
	}
}

type FlightPlanner struct {
	cc *country_mapper.CountryInfoClient
}

/*
Trip Assumptions:

An artist will fly from one gig to the next iff:

1. If the gigs are within two days of eachother
2. If the gigs outside the home base continent, on the same foreign continent, within two weeks of eachother

Otherwise we assume the artist returns home in between gigs.

*/

func withinTwoDays(d1, d2 time.Time) bool {
	days := d2.Sub(d1).Hours() / 24
	return days <= 2
}

func withinTwoWeeks(d1, d2 time.Time) bool {
	days := d2.Sub(d1).Hours() / 24
	return days <= 14
}

func sortByDate(events map[time.Time]event.Event) []time.Time {

	var keys []time.Time
	for k, _ := range events {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Before(keys[j])
	})
	return keys
}

func makeTrip(depCity, arrCity string, date time.Time) Trip {
	return Trip{
		DepCode: depCity,
		ArrCode: arrCity,
		Date:    date.String()[0:10],
	}
}

func (p FlightPlanner) sameForeignContinent(c1, c2, home string) bool {
	gigsSame := p.cc.MapByName(c1).Region == p.cc.MapByName(c1).Region
	homeData := p.cc.MapByName(home)
	c1Data := p.cc.MapByName(c1)
	if homeData != nil && c1Data != nil {
		homeDiff := p.cc.MapByName(home).Region != p.cc.MapByName(c1).Region
		return gigsSame && homeDiff
	}
	return false
}

func (p FlightPlanner) shouldFlyHome(e1, e2 event.Event, d1, d2 time.Time, homeCountry string) bool {
	if withinTwoDays(d1, d2) {
		return false
	}
	if withinTwoWeeks(d1, d2) && p.sameForeignContinent(e1.Country, e2.Country, homeCountry) {
		return false
	}
	return true

}

func (p FlightPlanner) Plan(a ra.Artist) (Trips, error) {
	var trips = make(Trips, 0)
	if a.AirCode == "" {
		return trips, errors.New("Cannot plan a trip without a home city.")
	}
	fmt.Printf("Creating flight plan..\n")
	homeCity, currCity := a.AirCode, a.AirCode
	events := sortByDate(a.Events)

	for index, date := range events {
		if index+1 == len(events) {
			trip := makeTrip(currCity, homeCity, date)
			trips = append(trips, trip)
			return trips, nil
		}
		event := a.Events[date]
		// Create a trip from the current city to the event we are looking at
		trip := makeTrip(currCity, event.AirCode, date)
		if trip.DepCode != trip.ArrCode {
			trips = append(trips, trip)
		}
		currCity = event.AirCode

		// Check the next event to see if we should then fly home
		nextEvent := a.Events[events[index+1]]
		if p.shouldFlyHome(event, nextEvent, date, events[index+1], a.Country) {
			homeTrip := makeTrip(currCity, homeCity, events[index+1])
			currCity = homeCity
			// Avoid tacking on a home trip from home
			if homeTrip.DepCode != homeTrip.ArrCode {
				trips = append(trips, homeTrip)
			}
		}

	}
	return trips, nil

}
