package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

func getEntryCount(db *sql.DB, from string, to string) int {
	var start string
	var end string

	if from == "" {
		start = "1970-01-01"
	} else {
		start = from
	}
	if to == "" {
		end = "2300-01-01"
	} else {
		end = to
	}

	q := fmt.Sprintf("SELECT ROWID FROM entries WHERE t >= '%s' AND t <= '%s 23:59:59:999999+02:00';", start, end)
	rows, err := db.Query(q)
	if err != nil {
		log.Fatal(err)
	}

	count := 0
	for rows.Next() {
		count += 1
	}

	return count
}

func readEntries(db *sql.DB, offset int, orderBy string, order string, from string, to string) ([]entry, error) {
	var start string
	var end string

	if from == "" {
		start = "1970-01-01"
	} else {
		start = from
	}
	if to == "" {
		end = "2300-01-01"
	} else {
		end = to
	}
	q := fmt.Sprintf("SELECT ROWID, * FROM entries WHERE t >= '%s' AND t <= '%s 23:59:59:999999+02:00' ORDER BY %s %s LIMIT 10 OFFSET %d;", start, end, orderBy, order, offset)
	rows, err := db.Query(q)
	if err != nil {
		log.Fatal(err)
	}
	var e []entry
	for rows.Next() {
		var en entry
		if err := rows.Scan(&en.Id, &en.Sys, &en.Dys, &en.Puls, &en.Sport,
			&en.T); err != nil {
			return e, err
		}
		e = append(e, entry{en.Id, en.Sys, en.Dys, en.Puls, en.Sport, en.T})
	}
	return e, nil
}

func readAllEntries(db *sql.DB) ([]entry, error) {
	q := fmt.Sprintf("SELECT ROWID, * FROM entries;")
	rows, err := db.Query(q)
	if err != nil {
		log.Fatal(err)
	}
	var e []entry
	for rows.Next() {
		var en entry
		if err := rows.Scan(&en.Id, &en.Sys, &en.Dys, &en.Puls, &en.Sport,
			&en.T); err != nil {
			return e, err
		}
		e = append(e, entry{en.Id, en.Sys, en.Dys, en.Puls, en.Sport, en.T})
	}
	return e, nil
}

func readEntry(db *sql.DB, id int) (entry, error) {
	q := fmt.Sprintf("SELECT ROWID, * FROM entries where ROWID = %d;", id)
	rows, err := db.Query(q)
	if err != nil {
		log.Fatal(err)
	}
	var e entry
	for rows.Next() {
		var en entry
		if err := rows.Scan(&en.Id, &en.Sys, &en.Dys, &en.Puls, &en.Sport,
			&en.T); err != nil {
			return e, err
		}
		e = entry{en.Id, en.Sys, en.Dys, en.Puls, en.Sport, en.T}
	}
	return e, nil
}

func writeEntry(e entry, db *sql.DB) {
	_, err := db.Exec("insert into entries values ($1, $2, $3, $4, $5);", e.Sys, e.Dys, e.Puls, e.Sport, e.T)
	if err != nil {
		log.Fatal(err)
	}
}

func updateEntry(e entry, db *sql.DB) {
	_, err := db.Exec("update entries set sys = $1, dys = $2, puls= $3, sport = $4 where ROWID = $5;", e.Sys, e.Dys, e.Puls, e.Sport, e.Id)
	if err != nil {
		log.Fatal(err)
	}
}

func generateData(e []entry) ([]string, []int64, []int64, []int64) {

	l := len(e)

	lab := make([]string, l)
	sys := make([]int64, l)
	dys := make([]int64, l)
	puls := make([]int64, l)

	for i, en := range e {
		lab[i] = en.T.Format(time.DateOnly)
		sys[i] = en.Sys
		dys[i] = en.Dys
		puls[i] = en.Puls
	}

	return lab, sys, dys, puls
}
