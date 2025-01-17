package handlers

import (
	"database/sql"
	"fmt"
)

type DBClient struct {
	*sql.DB
}

type Download struct {
	ServiceType string
	DownloadURL string
}

func (db *DBClient) AddDownloadRecord(email, serviceType, downloadURL string) error {
	query := `
		INSERT INTO downloads (user_email, file_type, signed_url, created_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err := db.Exec(query, email, serviceType, downloadURL)
	if err != nil {
		return fmt.Errorf("failed to insert download record: %w", err)
	}
	return nil
}

func (db *DBClient) FetchRecentDownloads(email string) ([]Download, error) {
	query := `
		SELECT file_type, signed_url
		FROM downloads
		WHERE user_email = $1 AND created_at >= NOW() - INTERVAL '10 minutes'
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query, email)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent downloads: %w", err)
	}
	defer rows.Close()

	var downloads []Download
	for rows.Next() {
		var download Download
		if err := rows.Scan(&download.ServiceType, &download.DownloadURL); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		downloads = append(downloads, download)
	}

	return downloads, nil
}
