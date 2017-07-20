package dwd

import (
	"fmt"
	"testing"
)

func TestStations(t *testing.T) {
	ss, err := Stations()
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range ss {
		fmt.Println(s)
	}
}

func TestForecast(t *testing.T) {
	var s Session
	s.Init()
	wfs, err := s.Forecast("06676")
	if err != nil {
		t.Fatal(err)
	}
	for _, wf := range wfs {
		fmt.Println(wf)
	}
}
