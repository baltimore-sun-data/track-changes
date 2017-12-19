package main

import (
	"context"
	"fmt"
	"log"
	"strings"

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

	a.Lock()
	// can't defer because of call to jobs().start()

	a.data, err = pageInfofromRows(a.data, sheet.Rows)
	if err != nil {
		a.Unlock()
		return err
	}

	a.title = spreadsheet.Properties.Title
	// Kill off existing jobs
	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
	}

	var ctx context.Context
	ctx, a.cancel = context.WithCancel(context.Background())
	a.Unlock()

	log.Printf("Succesfully processed Google Sheet")
	go a.jobs().start(ctx)

	return nil
}

func pageInfofromRows(oldData []pageInfo, rows [][]spreadsheet.Cell) (pages []pageInfo, err error) {
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

	// Save info between refreshes
	oldInfo := map[string]*pageInfo{}
	for i := range oldData {
		pp := &oldData[i]
		oldInfo[pp.Id] = pp
	}

	for _, row := range rows[1:] {
		if len(row) != rowLen {
			return nil, fmt.Errorf("malformed row")
		}

		// Number of cells returned may be much bigger than what is used
		// So just bail if you hit an empty space
		if row[0].Value == "" {
			return
		}

		// Use old info as basis
		pi := oldInfo[getTrimmed(row, idIdx)]
		if pi == nil {
			pi = &pageInfo{}
		}
		pi.Id = getTrimmed(row, idIdx)
		pi.HomePageUrl = getTrimmed(row, homepageUrlIdx)
		pi.Twitter = getTrimmed(row, screennameIdx)
		pi.DisplayName = getTrimmed(row, nameIdx)
		pi.Url = getTrimmed(row, notificationUrlIdx)
		pi.Selector = getTrimmed(row, selectorIdx)

		pages = append(pages, *pi)
	}

	return
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

func getTrimmed(row []spreadsheet.Cell, idx int) string {
	return strings.TrimSpace(row[idx].Value)
}
