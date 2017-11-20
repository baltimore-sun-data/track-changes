package main

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"gopkg.in/Iwark/spreadsheet.v2"
)

func fromSheet(gdoc string) error {
	keydata, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		return errors.WithMessage(err, "could not read credentials")
	}

	conf, err := google.JWTConfigFromJSON(keydata, spreadsheet.Scope)
	if err != nil {
		return errors.WithMessage(err, "could not parse credentials")
	}

	client := conf.Client(context.Background())
	service := spreadsheet.NewServiceWithClient(client)
	spreadsheet, err := service.FetchSpreadsheet(gdoc)
	if err != nil {
		return errors.WithMessage(err, "failure getting Google Sheet")
	}

	sheet, err := spreadsheet.SheetByIndex(0)
	if err != nil {
		return errors.WithMessage(err, "Sheet does not contain expected data")
	}

	if err = data.fromRows(sheet.Rows); err != nil {
		return err
	}

	return nil
}

func (a *apiResponse) fromRows(rows [][]spreadsheet.Cell) error {
	a.Lock()
	defer a.Unlock()

	if len(rows) < 1 {
		return fmt.Errorf("Google Sheet does not contain any rows")
	}

	rowLen := len(rows[0])

	var (
		idIdx, nameIdx, homepageUrlIdx, notificationUrlIdx int
		selectorIdx, screennameIdx                         int
	)

	if err := indexFields(rows[0], map[string]*int{
		"id":                 &idIdx,
		"name":               &nameIdx,
		"homepage_url":       &homepageUrlIdx,
		"notification_url":   &notificationUrlIdx,
		"selector":           &selectorIdx,
		"twitter_screenname": &screennameIdx,
	}); err != nil {
		return errors.WithMessage(err, "spreadsheet missing header")
	}

	for i, row := range rows {
		if i == 0 {
			continue
		}

		if len(row) != rowLen {
			return fmt.Errorf("malformed row")
		}

		if row[0].Value == "" {
			return nil
		}

		data.data = append(data.data, jsonData{
			Id:          row[idIdx].Value,
			HomePageUrl: row[homepageUrlIdx].Value,
			Twitter:     row[screennameIdx].Value,
			DisplayName: row[nameIdx].Value,
			Url:         row[notificationUrlIdx].Value,
			Selector:    row[selectorIdx].Value,
		})
	}

	return nil
}

func indexFields(row []spreadsheet.Cell, fields map[string]*int) error {
	// Initialize pointer values to sentinal
	for _, p := range fields {
		*p = -1
	}

	for i, cell := range row {
		if p, ok := fields[cell.Value]; ok {
			*p = i
		}
	}

	// Check for unset header
	for header, p := range fields {
		if *p == -1 {
			return fmt.Errorf("could not find header %q", header)
		}
	}
	return nil
}
