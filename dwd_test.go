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

func TestOverview(t *testing.T) {
	var s Session
	s.Init()
	wfs, err := s.Overview("06676")
	if err != nil {
		t.Fatal(err)
	}
	for _, wf := range wfs {
		fmt.Println(wf)
	}
}
