package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	dir := flag.String("path", "migrations", "migrations directory (relative to cwd)")
	cmd := flag.String("cmd", "up", "up | down | version")
	flag.Parse()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	m, err := migrate.New(fmt.Sprintf("file://%s", *dir), dsn)
	if err != nil {
		log.Fatalf("migrate new: %v", err)
	}
	defer m.Close()

	switch *cmd {
	case "up":
		err = m.Up()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(err)
		}
	case "down":
		err = m.Down()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(err)
		}
	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("version=%d dirty=%v\n", v, dirty)
	default:
		log.Fatal("unknown -cmd")
	}
}
