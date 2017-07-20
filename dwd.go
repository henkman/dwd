package dwd

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"time"
)

var (
	stations []Station
)

type Station struct {
	Pk             string
	Name           string
	X              float64
	Y              float64
	Altitude       int
	Priority       int
	Private        bool
	HasMeasurement bool
	HasWarnregion  bool
	Country        string
	Active         bool
}

type Direction int

func (d *Direction) Radian() float32 {
	return float32(float64(*d) * (math.Pi / 180))
}

func (d *Direction) Degree() int {
	return int(*d)
}

type Forecast struct {
	WindGust       int
	WindSpeed      int
	DayDate        time.Time
	WindDirection  Direction
	Precipitation  float32
	Icon2          int
	Icon1          int
	TemperatureMin float32
	TemperatureMax float32
}

func Stations() ([]Station, error) {
	if len(stations) != 0 {
		return stations, nil
	}
	fd, err := os.Open("./stations.csv")
	if err != nil {
		return nil, err
	}
	records, err := csv.NewReader(fd).ReadAll()
	fd.Close()
	if err != nil {
		return nil, err
	}
	s := make([]Station, 0, len(records[1:]))
	for _, record := range records[1:] {
		x, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return nil, err
		}
		y, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			return nil, err
		}
		altitude, err := strconv.ParseInt(record[4], 10, 32)
		if err != nil {
			return nil, err
		}
		priority, err := strconv.ParseInt(record[5], 10, 32)
		if err != nil {
			return nil, err
		}
		private, err := strconv.ParseInt(record[6], 10, 8)
		if err != nil {
			return nil, err
		}
		hasMeasurement, err := strconv.ParseInt(record[7], 10, 8)
		if err != nil {
			return nil, err
		}
		hasWarnregion, err := strconv.ParseInt(record[8], 10, 8)
		if err != nil {
			return nil, err
		}
		active, err := strconv.ParseInt(record[10], 10, 8)
		if err != nil {
			return nil, err
		}
		s = append(s, Station{
			Pk:             record[0],
			Name:           record[1],
			X:              x,
			Y:              y,
			Altitude:       int(altitude),
			Priority:       int(priority),
			Private:        private == 1,
			HasMeasurement: hasMeasurement == 1,
			HasWarnregion:  hasWarnregion == 1,
			Country:        record[9],
			Active:         active == 1,
		})
	}
	stations = s
	return stations, nil
}

type Session struct {
	cli http.Client
}

func (s *Session) Init() error {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	s.cli.Jar = jar
	return nil
}

func (s *Session) IsInitialized() bool {
	return s.cli.Jar != nil
}

func (s *Session) request(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "dwd api")
	return s.cli.Do(req)
}

func (s *Session) Overview(station string) ([]Forecast, error) {
	ru := fmt.Sprintf(
		"https://app-prod-ws.warnwetter.de/v16/stationOverview?stationIds=%s",
		station)
	res, err := s.request("GET", ru, nil)
	if err != nil {
		return nil, err
	}
	var result map[string][]struct {
		WindGust       int    `json:"windGust"`
		WindSpeed      int    `json:"windSpeed"`
		DayDate        string `json:"dayDate"`
		WindDirection  int    `json:"windDirection"`
		Precipitation  int    `json:"precipitation"`
		Icon2          int    `json:"icon2"`
		Icon1          int    `json:"icon1"`
		TemperatureMin int    `json:"temperatureMin"`
		TemperatureMax int    `json:"temperatureMax"`
	}
	err = json.NewDecoder(res.Body).Decode(&result)
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	for k := range result {
		rk := result[k]
		fcr := make([]Forecast, len(rk))
		for i, r := range rk {
			date, err := time.Parse("2006-01-02", r.DayDate)
			if err != nil {
				return nil, err
			}
			fcr[i] = Forecast{
				WindGust:       r.WindGust / 10,
				WindSpeed:      r.WindSpeed / 10,
				DayDate:        date,
				WindDirection:  Direction((r.WindDirection/10 + 180) % 360),
				Precipitation:  float32(r.Precipitation) / 10,
				Icon2:          r.Icon1,
				Icon1:          r.Icon2,
				TemperatureMin: float32(r.TemperatureMin) / 10,
				TemperatureMax: float32(r.TemperatureMax) / 10,
			}
		}
		return fcr, nil
	}
	return nil, errors.New("answer didn't contain forecast entries")
}
