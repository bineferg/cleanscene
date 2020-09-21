package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cleanscene.flights/lib/airports"
	"github.com/cleanscene.flights/lib/atmos"
	"github.com/cleanscene.flights/lib/crawler"
	"github.com/cleanscene.flights/lib/flight"
	"github.com/cleanscene.flights/lib/google"
	"github.com/cleanscene.flights/lib/ra"
	country_mapper "github.com/pirsquare/country-mapper"
)

var (
	tourYear   = flag.String("tour.year", "2019", "year in which to scrape artist event schedule")
	outputDir  = flag.String("output.dir", "./done/artist-pages", "directory to write flight data csv output to")
	artistFile = flag.String("artist.inputs", "./fixtures/ra-1000.csv", "precompiled, editied list of the RA artists")

	googleApiKey  = flag.String("google.apikey", os.Getenv("GOOGLE_API_KEY"), "google api key for airports svc")
	atmosAcctID   = flag.String("atmos.acctID", os.Getenv("ATMOS_ACCOUNT_ID"), "account id for atmosfaire api")
	atmosPassword = flag.String("atmos.pass", os.Getenv("ATMOS_PASSWORD"), "password for atmosfaire api")
	edgeApiKey    = flag.String("edge.apiKey", os.Getenv("EDGE_API_KEY"), "key for edge api to find nearst airport code")
)

// We begin by crawling the RA top 1000 artists.
const baseUrl = "https://www.residentadvisor.net"

// Atmosfair api url use for flight emission calculations.
const atmosUrl = "https://api.atmosfair.de/api/emission/flight"

// For every event for each artist, we record these data points.
var metaDataHeaders = []string{"DEPARTURE", "ARRIVAL", "DATE", "OFFSET", "CARBON OUTPUT", "FUEL", "DISTANCE"}

// By default, dont fail on error simply log.
var errCheck = func(err error) {
	if err != nil {
		log.Info(err.Error())
	}
}

// Write flight & carbon output data to artist csv file
func writeTo(outputs []atmos.Output, artistName, outputDir string) {
	csvfile, err := os.Create(fmt.Sprintf("%s/%s.csv", outputDir, artistName))

	if err != nil {
		fmt.Printf("failed creating file: %s", err)
		return
	}

	csvwriter := csv.NewWriter(csvfile)
	csvwriter.Write(metaDataHeaders)
	for _, output := range outputs {
		row := []string{
			output.DepartCode,
			output.ArrivalCode,
			output.FlightDay,
			fmt.Sprintf("€%f", output.OffsetEuros),
			fmt.Sprintf("%f kg", output.CarbonOutput),
			fmt.Sprintf("%f L", output.FuelInLiter),
			fmt.Sprintf("%d km", output.Distance),
		}
		err = csvwriter.Write(row)
		errCheck(err)
	}

	csvwriter.Flush()

	csvfile.Close()
}

func main() {
	parseFlags()
	djCrawler, err := crawler.New(baseUrl)
	if err != nil {
		log.Fatal(err)
	}
	googleApi := google.NewApi(*googleApiKey)
	airSvc, err := airports.New("./airports.csv", *edgeApiKey, googleApi)
	if err != nil {
		log.Fatal(err)
	}

	raSvc := ra.New(airSvc, djCrawler, *outputDir, *tourYear)
	cclient, err := country_mapper.Load()
	if err != nil {
		log.Fatal(err)
	}

	planner := flight.NewPlanner(cclient)
	atmosSvc := atmos.NewFair(atmosUrl, *atmosAcctID, *atmosPass)
	artists, err := raSvc.LoadArtists(*artistFile)
	if err != nil {
		log.Fatal(err)
	}

	for _, artist := range artists {
		events, err := raSvc.LoadEvents(artist)
		errCheck(err)
		artist.Events = events
		trips, err := planner.Plan(artist)
		errCheck(err)
		outputs, err := atmosSvc.Calculate(trips)
		errCheck(err)
		writeTo(outputs, artist.Name, *outputDir)
	}

}

func parseFlags() {
	flag.Parse()
	if *tourYear == "" {
		log.Fatal("missing tour year for aritst")
	}
	if *outputDir == "" {
		log.Fatal("outputdir missing to write files to")
	}
	if *artistFile == "" {
		log.Fatal("missing pre-compiled list of artists intended to scrape")
	}
	if *googleApiKey == "" {
		log.Fatal("missing googlepai key to find nearest airport")
	}
	if *atmosAcctID == "" {
		log.Fatal("atmosfaire account id for carbon emissions api")
	}
	if *atmosPass == "" {
		log.Fatal("atmosfaire password for carbon emissions api")
	}
	if *edgeApiKey == "" {
		log.Fatal("edge api key missing for nearest aircode")
	}

}
