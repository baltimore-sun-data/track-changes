package main

import (
	"sync"
)

type dataStore struct {
	m sync.Map
}

func (d *dataStore) get(sheet string) (ar *apiResponse, err error) {
	// Try to get an existing sheet (common case)
	// If we can't get one, try to create one
	// If we succeed in creating one, let us know if it worked or not
	// If we got stuck in a race, return the once created first in the race
	a, ok := d.m.Load(sheet)
	if ok {
		ar = a.(*apiResponse)
	} else {
		ar = &apiResponse{}
		a, loaded := d.m.LoadOrStore(sheet, ar)
		if !loaded {
			err = ar.fromSheet(sheet)
		} else {
			ar = a.(*apiResponse)
		}
	}
	return
}

func (d *dataStore) refresh(sheet string) error {
	var ar *apiResponse
	a, ok := d.m.Load(sheet)
	if ok {
		ar = a.(*apiResponse)
	} else {
		ar = &apiResponse{}
		a, loaded := d.m.LoadOrStore(sheet, ar)
		if loaded {
			ar = a.(*apiResponse)
		}
	}

	return ar.fromSheet(sheet)
}
