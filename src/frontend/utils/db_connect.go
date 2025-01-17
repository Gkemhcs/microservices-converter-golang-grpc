package utils

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	
)

func ConnectToDB(logger *logrus.Logger) (*sql.DB, error) {
	// Fetch database connection details from environment variables with POSTGRES_ prefix
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")

	// Check for missing environment variables
	if dbUser == "" || dbPassword == "" || dbName == "" || dbHost == "" || dbPort == "" {
		logger.Error("Missing required database environment variables")
		return nil, fmt.Errorf("missing required database environment variables")
	}

	// Create connection string
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		dbUser, dbPassword, dbName, dbHost, dbPort)

	// Connect to the database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Errorf("Failed to connect to database: %v", err)
		return nil, err
	}

	// Test the database connection
	err = db.Ping()
	if err != nil {
		logger.Errorf("Failed to ping database: %v", err)
		return nil, err
	}

	logger.Info("Successfully connected to the database")
	return db, nil
}
