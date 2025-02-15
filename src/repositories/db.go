package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/cinar/indicator/v2/asset"
	"github.com/cinar/indicator/v2/helper"
	"pivetta.se/crypro-spotter/src/utils"
)

func GetLatestSnapshot(db *sql.DB, a string) (*asset.Snapshot, error) {
	query := `SELECT date, open, high, low, close, volume FROM snapshots WHERE asset = $1 ORDER BY date DESC LIMIT 1`
	layout := "2006-01-02T15:04:05Z"

	row := db.QueryRow(query, a)
	var dateStr string
	var snapshot asset.Snapshot
	err := row.Scan(&dateStr, &snapshot.Open, &snapshot.High, &snapshot.Low, &snapshot.Close, &snapshot.Volume)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getMostRecentSnapshot: %w", err)
	}

	localLocation := time.Now().Location()
	parsed, err := time.ParseInLocation(layout, dateStr, localLocation)
	if err != nil {
		return nil, fmt.Errorf("getMostRecentSnapshot, parse date: %w", err)
	}

	snapshot.Date = parsed

	return &snapshot, nil
}

func GetSnapshots(db *sql.DB, a string, limit int) ([]*asset.Snapshot, error) {
	query := `SELECT date, open, high, low, close, volume FROM snapshots WHERE asset = $1 ORDER BY date DESC LIMIT $2`
	layout := "2006-01-02T15:04:05Z"
	var dateStr string

	rows, err := db.Query(query, a, limit)
	if err != nil {
		return nil, fmt.Errorf("getSnapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []*asset.Snapshot
	for rows.Next() {
		var snapshot asset.Snapshot
		err := rows.Scan(&dateStr, &snapshot.Open, &snapshot.High, &snapshot.Low, &snapshot.Close, &snapshot.Volume)
		if err != nil {
			return nil, fmt.Errorf("getSnapshots: %w", err)
		}

		localLocation := time.Now().Location()
		parsed, err := time.ParseInLocation(layout, dateStr, localLocation)
		if err != nil {
			return nil, fmt.Errorf("getSnapshots, parse date: %w", err)
		}

		snapshot.Date = parsed
		snapshots = append(snapshots, &snapshot)
	}

	return snapshots, nil
}

func InsertSnapshots(db *sql.DB, a string, ss chan *asset.Snapshot) error {
	query := `INSERT INTO snapshots (asset, date, open, high, low, close, volume) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	for snapshot := range ss {
		// log.Printf("Inserting snapshot: %+v\n", snapshot)
		_, err := db.Exec(query, a, snapshot.Date, snapshot.Open, snapshot.High, snapshot.Low, snapshot.Close, snapshot.Volume)
		if err != nil {
			return fmt.Errorf("insertSnapshots: %w", err)
		}
	}
	return nil
}

func Cleanup(db *sql.DB, a string) error {
	date := time.Now().Add(-15 * 24 * time.Hour).Truncate(24 * time.Hour).UTC()
	log.Printf("Cleaning up snapshots before %v\n", date)
	query := `DELETE FROM snapshots WHERE asset = $1 AND date < $2`
	_, err := db.Exec(query, a, date)
	if err != nil {
		return fmt.Errorf("cleanup: %w", err)
	}
	return nil
}

func NewDBRepository(a string, limit int) (asset.Repository, error) {
	db := utils.GetDb()
	repo := asset.NewInMemoryRepository()

	ss, err := GetSnapshots(db, a, limit)
	if err != nil {
		return nil, fmt.Errorf("error fetching snapshots: %w", err)
	}

	sschan := helper.SliceToChan(ss)
	err = repo.Append(a, sschan)
	if err != nil {
		return nil, fmt.Errorf("error appending snapshots: %w", err)
	}

	return repo, nil
}
