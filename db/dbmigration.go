package dbmigration

import (
	"database/sql"
	"log"

	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate(conn *sql.DB) {
	log.Println("Database migration start")

	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		log.Fatalf("could not start migration: %s", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://db/migrations", "postgres", driver)

	if err != nil {
		log.Fatalf("database migration failed: %s", err)
	}

	// run migration down
	if err := m.Down(); err != nil {
		log.Fatalf("database migration down failed: %s", err)
	}

	// run migration up
	if err := m.Up(); err != nil {
		log.Fatalf("database migration up failed: %s", err)
	}

	log.Println("database migration complete")
}
