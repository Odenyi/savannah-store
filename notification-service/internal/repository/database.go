package repository

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"savannah-store/notification-service/internal/constants"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

func DbInstance(dbname string) *sql.DB {

	host := os.Getenv("DB_HOST")
	password := os.Getenv("DB_PASSWORD")
	port := os.Getenv("DB_PORT")
	username := os.Getenv("DB_USERNAME")

	dbURI := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&multiStatements=true", username, password, host, port, dbname, "utf8")
	Db, err := sql.Open("mysql", dbURI)
	checkErr(err)

	err = Db.Ping()
	checkErr(err)
	return Db
}
func CheckConnectionStatus() (int, interface{}) {

	res := make(map[string]interface{})

	db := DbInstance(os.Getenv("IDENTITY_DB_NAME"))

	defer db.Close()

	err := db.Ping()
	if err == nil {

		res["db_status"] = "sent successful ping"

	} else {

		res["db_status"] = err.Error()
		return http.StatusInternalServerError, res
	}

	return http.StatusOK, res

}
func checkErr(err error) {

	if err != nil {
		logrus.WithFields(logrus.Fields{constants.DESCRIPTION: "got error connecting to db"}).Error(err.Error())

	}
}
