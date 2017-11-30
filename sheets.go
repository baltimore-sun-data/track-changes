package main

import (
	"context"
	"fmt"
	"log"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"gopkg.in/Iwark/spreadsheet.v2"
)

var sheetsClientSecret = MustGetEnv("GOOGLE_CLIENT_SECRET")

func (a *apiResponse) fromSheet(gdoc string) error {
	log.Printf("Connecting to Google Sheets for %q", gdoc)

	conf, err := google.JWTConfigFromJSON([]byte(sheetsClientSecret), spreadsheet.Scope)
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

	ctx, err := a.fromRows(sheet.Rows)
	if err != nil {
		return err
	}

	log.Printf("Succesfully processed Google Sheet")
	go a.jobs().start(ctx)

	return nil
}

func (a *apiResponse) fromRows(rows [][]spreadsheet.Cell) (ctx context.Context, err error) {
	a.Lock()
	defer a.Unlock()

	if len(rows) < 1 {
		return nil, fmt.Errorf("Google Sheet does not contain any rows")
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
		return nil, errors.WithMessage(err, "spreadsheet missing header")
	}

	// Kill off existing jobs
	if a.cancel != nil {
		a.cancel()
	}

	ctx, a.cancel = context.WithCancel(context.Background())
	a.data = nil

	for _, row := range rows[1:] {
		if len(row) != rowLen {
			return nil, fmt.Errorf("malformed row")
		}

		if row[0].Value == "" {
			return ctx, nil
		}

		a.data = append(a.data, pageInfo{
			Id:          row[idIdx].Value,
			HomePageUrl: row[homepageUrlIdx].Value,
			Twitter:     row[screennameIdx].Value,
			DisplayName: row[nameIdx].Value,
			Url:         row[notificationUrlIdx].Value,
			Selector:    row[selectorIdx].Value,
		})
	}

	return ctx, nil
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
