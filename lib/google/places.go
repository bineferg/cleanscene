package google

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Places interface {
	QueryCoordinates(string, QueryType) (float64, float64, error)
}

type Api struct {
	key           string
	hostCoordUrl  string
	hostCoordText string
}

func NewApi(key string) Places {
	return Api{
		key:           key,
		hostCoordUrl:  "https://maps.googleapis.com/maps/api/place/textsearch/json?query=",
		hostCoordText: "https://maps.googleapis.com/maps/api/place/findplacefromtext/json?input=",
	}
}

type PlaceResp struct {
	Results    []Result `json:"results"`
	Candidates []Result `json:"candidates"`
}
type Result struct {
	Name     string   `json:"name"`
	Geometry Geometry `json:"geometry"`
}

type Geometry struct {
	Location Location `json:"location"`
}

type Location struct {
	Long float64 `json:"lng"`
	Lat  float64 `json:"lat"`
}

type QueryType int

const (
	CoordQuery QueryType = iota
	TextQuery
)

var (
	coordParams = "inputtype=textquery&fields=name,formatted?address,geometry&key="
	textParams  = "inputtype=textquery&fields=geometry&key="
)

func (api Api) QueryCoordinates(place string, queryType QueryType) (float64, float64, error) {
	var (
		rr    = PlaceResp{}
		query string
	)
	switch queryType {
	case CoordQuery:
		query = fmt.Sprintf("%s%s&%s%s", api.hostCoordUrl, coordParams, place, api.key)
	case TextQuery:
		place = strings.ReplaceAll(place, ",", "")
		query = fmt.Sprintf("%s%s&%s%s", api.hostCoordText, textParams, url.PathEscape(place), api.key)
	}
	resp, err := http.Get(query)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&rr)

	// For simplicity we choose the first response of which ever one is populated.
	if len(rr.Candidates) != 0 {
		return rr.Candidates[0].Geometry.Location.Long, rr.Candidates[0].Geometry.Location.Lat, nil

	}
	if len(rr.Results) != 0 {
		return rr.Results[0].Geometry.Location.Long, rr.Results[0].Geometry.Location.Lat, nil

	}
	return 0, 0, errors.New("no coordintes found!")

}
