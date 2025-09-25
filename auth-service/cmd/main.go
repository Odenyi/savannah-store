package main

import (
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"

	"savannah-store/auth-service/internal/repository"

	_ "savannah-store/auth-service/docs"

	"fmt"
	app "savannah-store/auth-service/internal/handlers"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"path/filepath"
	"runtime"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println(" No .env file found, using system environment variables")
	}

	// setup database
	dbInstance := repository.DbInstance(os.Getenv("AUTH_DB_NAME"))

	driver, err := mysql.WithInstance(dbInstance, &mysql.Config{})
	if err != nil {

		panic(err)
	}

	migrationPath := fmt.Sprintf("file://%s/../migrations", GetRootPath())

	m, err := migrate.NewWithDatabaseInstance(migrationPath, "mysql", driver)
	if err != nil {

		log.Printf("migration setup error %s ",err.Error())
	}

	err = m.Up() // or m.Step(2) if you want to explicitly set the number of migrations to run
	if err != nil {

		log.Printf("migration error %s ",err.Error())
	}

	router := &app.App{}

	fmt.Println(" Init Routers V3")
	router.Initialize()
	router.Run()
}

func GetRootPath() string {

	_, b, _, _ := runtime.Caller(0)

	// Root folder of this project
	return filepath.Join(filepath.Dir(b), "./")
}