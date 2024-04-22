package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/theLemionday/upload-photos-backend/informations"
)

var (
	db      *sql.DB
	queries informations.Queries
)

func init() {
	// ctx := context.Background()

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, hostname, port, database))
	if err != nil {
		log.Fatal(err)
	}

	queries = *informations.New(db)
}

func CreateImageInformation(name string) error {
	_, err := queries.CreateImageInformation(context.Background(), informations.CreateImageInformationParams{
		Name:    name,
		Created: time.Now(),
	})
	if err != nil {
		return err
	}

	return nil
}
