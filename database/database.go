package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func DBConnect() (*sql.DB, error) {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	dbHost := os.Getenv("DB_HOST")
	dbConn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&tls=true&interpolateParams=true",
		dbUser,
		dbPassword,
		dbHost,
		dbPort,
		dbName,
	)

	db, err := sql.Open("mysql", dbConn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	fmt.Println("database connected!")
	return db, nil
}
