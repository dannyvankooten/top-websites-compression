package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite3", "data.sqlite3?_timeout=5000")
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
}

type Site struct {
	Url            string
	Rank           int
	Compr          string
	Size           int
	SizeCompressed int
	CacheLifetime  int
}

func (s *Site) Savings() int {
	return (s.Size - s.SizeCompressed) / 1024
}

func querySites(query string) ([]Site, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	results := make([]Site, 0, 128)
	for rows.Next() {
		var s Site
		if err := rows.Scan(&s.Url, &s.Rank, &s.Compr, &s.Size, &s.SizeCompressed); err != nil {
			return nil, err
		}

		results = append(results, s)
	}

	return results, nil
}

func getSites() ([]Site, error) {
	return querySites(`
		SELECT url, rank, compression, size, size_compressed
		FROM sites
		WHERE checked_at IS NULL
			OR checked_at < DATETIME('now', '-1 month')
	`)
}

func getSitesOrderedBySavings() ([]Site, error) {
	return querySites(`
		SELECT url, rank, compression, size, size_compressed
		FROM sites
		WHERE checked_at IS NOT NULL
		ORDER BY (size - size_compressed) DESC
		LIMIT 0, 500
	`)
}

func updateSite(s *Site) error {
	_, err := db.Exec(`
		UPDATE sites SET checked_at = DATETIME('now'), compression = ?, size = ?, size_compressed = ? WHERE url = ?
	`, s.Compr, s.Size, s.SizeCompressed, s.Url)
	return err
}
