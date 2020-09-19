package atmos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cleanscene.flights/lib/flight"
)

type AtmosFair interface {
	Calculate(trips []flight.Trip) ([]Output, error)
}

type service struct {
	acctID   string
	password string
	host     string
	cli      *http.Client
}

func NewFair(host, acctID, password string) AtmosFair {
	return service{acctID: acctID, password: password, host: host, cli: &http.Client{}}
}

type AtmosResp struct {
	Status      string       `json:"status"`
	Errors      []string     `json:"errors"`
	Carbon      float64      `json:"co2"`
	OffsetInEu  float64      `json:"offsetInEUR"`
	FuelInLiter float64      `json:"fuelInLiter"`
	Distance    int          `json:"distance"`
	Flights     []FlightResp `json:"flights"`
}

type FlightResp struct {
	CarbonOutput  float64 `json:"co2"`
	FuelInLiter   float64 `json:"fuelInLiter"`
	Distance      int     `json:"distance"`
	OffsetInEu    float64 `json:"offsetInEUR"`
	DepartureDate string  `json:"departureDate"`
	DepartCode    string  `json:"departure"`
	ArrivalCode   string  `json:"arrival"`
}

type AtmosReq struct {
	AccountID string   `json:"accountId"`
	Password  string   `json:"password"`
	Flights   []Flight `json:"flights"`
}

type Flight struct {
	DepartCode    string `json:"departure"`
	ArrivalCode   string `json:"arrival"`
	PassCount     int    `json:"passengerCount"`
	DepartureDate string `json:"departureDate"`
	FlightCount   int    `json:"flightCount"`
}

func (s service) Calculate(trips []flight.Trip) ([]Output, error) {
	var outputs = make([]Output, 0)
	atmosReq := AtmosReq{
		AccountID: s.acctID,
		Password:  s.password,
		Flights:   make([]Flight, 0),
	}
	for _, trip := range trips {
		if trip.DepCode == "" || trip.ArrCode == "" {
			continue
		}
		atmosReq.Flights = append(atmosReq.Flights, Flight{
			DepartCode:    trip.DepCode,
			ArrivalCode:   trip.ArrCode,
			DepartureDate: trip.Date,
			FlightCount:   1,
			PassCount:     1,
		})
	}
	resp, err := s.bulkReq(atmosReq)
	if err != nil {
		return outputs, err
	}
	finalFlights := s.retryAndMerge(resp.Flights)
	for _, flight := range finalFlights {
		outputs = append(outputs, Output{
			ArrivalCode:  flight.ArrivalCode,
			DepartCode:   flight.DepartCode,
			FlightDay:    flight.DepartureDate,
			OffsetEuros:  flight.OffsetInEu,
			CarbonOutput: flight.CarbonOutput,
			FuelInLiter:  flight.FuelInLiter,
			Distance:     flight.Distance,
		})
	}
	return outputs, nil
}
func (s service) bulkReq(req AtmosReq) (AtmosResp, error) {
	var atmosResp AtmosResp
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(req)
	httpReq, _ := http.NewRequest("POST", s.host, b)
	httpReq.Header.Set("Accept", "application/json, text/plain, */*")
	httpReq.Header.Set("Content-Type", "application/json;charset=UTF-8")
	resp, err := s.cli.Do(httpReq)
	defer resp.Body.Close()
	if err != nil {
		return atmosResp, err
	}
	json.NewDecoder(resp.Body).Decode(&atmosResp)
	if atmosResp.Status != "SUCCESS" {
		fmt.Println(atmosResp.Errors[0])
	}
	return atmosResp, nil
}

func (s service) syncReq(req AtmosReq) []FlightResp {
	var fResp = make([]FlightResp, 0)
	for _, flight := range req.Flights {
		rr := AtmosReq{
			AccountID: s.acctID,
			Password:  s.password,
			Flights:   []Flight{flight},
		}
		var atmosResp = AtmosResp{}
		b := new(bytes.Buffer)
		json.NewEncoder(b).Encode(rr)
		httpReq, _ := http.NewRequest("POST", s.host, b)
		httpReq.Header.Set("Accept", "application/json, text/plain, */*")
		httpReq.Header.Set("Content-Type", "application/json;charset=UTF-8")
		resp, err := s.cli.Do(httpReq)
		defer resp.Body.Close()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		json.NewDecoder(resp.Body).Decode(&atmosResp)
		if atmosResp.Status != "SUCCESS" {
			fmt.Printf("ATMOS ERROR: %v for request: %v\n", atmosResp.Errors, rr)
			continue
		}
		fResp = append(fResp, atmosResp.Flights[0])
	}
	return fResp

}

type Output struct {
	DepartCode   string
	ArrivalCode  string
	FlightDay    string
	OffsetEuros  float64
	CarbonOutput float64
	FuelInLiter  float64
	Distance     int
}

func findNullData(flights []FlightResp) (int, int) {
	var (
		firstEmptyFlightIdx = 0
		endEmptyFlightIdx   = len(flights) - 1
	)
	for index, flight := range flights {
		if flight.OffsetInEu == 0 {
			firstEmptyFlightIdx = index
			break
		}
	}
	for index, flight := range flights[firstEmptyFlightIdx:] {
		if flight.OffsetInEu != 0 {
			endEmptyFlightIdx = index
			break
		}
	}
	if firstEmptyFlightIdx >= endEmptyFlightIdx {
		return 0, 0
	}
	return firstEmptyFlightIdx, endEmptyFlightIdx

}

// Seems to be some sort of rate limit or bug with atmosfair, this handles that.
func (s service) retryAndMerge(firstAttempt []FlightResp) []FlightResp {
	start, end := findNullData(firstAttempt)
	retryFlights := make([]Flight, 0)
	emptyFlights := firstAttempt[start:end]
	for _, flight := range emptyFlights {
		retryFlights = append(retryFlights, Flight{
			ArrivalCode:   flight.ArrivalCode,
			DepartCode:    flight.DepartCode,
			DepartureDate: flight.DepartureDate,
			FlightCount:   1,
			PassCount:     1,
		})

	}
	syncResp := s.syncReq(AtmosReq{AccountID: s.acctID, Password: s.password, Flights: retryFlights})
	finalFlights := append(firstAttempt[:start], syncResp...)
	finalFlights = append(finalFlights, firstAttempt[end:]...)
	return finalFlights

}
