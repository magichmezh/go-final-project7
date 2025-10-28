package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const dateFormat = "20060102"

func afterNow(d1, d2 time.Time) bool {
	d1 = time.Date(d1.Year(), d1.Month(), d1.Day(), 0, 0, 0, 0, time.UTC)
	d2 = time.Date(d2.Year(), d2.Month(), d2.Day(), 0, 0, 0, 0, time.UTC)
	return d1.After(d2)
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("repeat rule is empty")
	}

	startDate, err := time.ParseInLocation(dateFormat, dstart, now.Location())
	if err != nil {
		return "", fmt.Errorf("invalid start date format: %w", err)
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", errors.New("repeat rule is empty")
	}

	rule := parts[0]
	arg := ""
	if len(parts) > 1 {
		arg = strings.Join(parts[1:], " ")
	}

	switch rule {
	case "d":
		if arg == "" {
			return "", errors.New("missing interval for d rule")
		}
		days, err := strconv.Atoi(arg)
		if err != nil {
			return "", errors.New("invalid day interval format")
		}
		if days < 1 || days > 400 {
			return "", errors.New("day interval out of allowed range (1-400)")
		}
		nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, now.Location())

		if startDate.After(nowDate) {
			return startDate.Format(dateFormat), nil
		}

		diffDays := int(nowDate.Sub(startDate).Hours() / 24)
		intervalsPassed := diffDays / days
		next := startDate.AddDate(0, 0, (intervalsPassed+1)*days)
		return next.Format(dateFormat), nil

	case "y":
		nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, now.Location())

		yearsPassed := nowDate.Year() - startDate.Year()
		candidate := startDate.AddDate(yearsPassed, 0, 0)
		if !candidate.After(nowDate) {
			yearsPassed++
		}
		next := startDate.AddDate(yearsPassed, 0, 0)
		return next.Format(dateFormat), nil

	default:
		return "", fmt.Errorf("unsupported repeat rule: %s", rule)
	}
}

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowParam := r.FormValue("now")
	dateParam := r.FormValue("date")
	repeatParam := r.FormValue("repeat")

	if dateParam == "" || repeatParam == "" {
		http.Error(w, "missing required parameters: date and repeat", http.StatusBadRequest)
		return
	}

	var now time.Time
	var err error
	if nowParam == "" {
		now = time.Now()
	} else {
		now, err = time.Parse(dateFormat, nowParam)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid now parameter: %v", err), http.StatusBadRequest)
			return
		}
	}

	next, err := NextDate(now, dateParam, repeatParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(next))
}
