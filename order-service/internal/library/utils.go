package library

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)
func GetValuesOnly(c echo.Context) (payload map[string]interface{},httpStatus int, err error ) {

	return GetJSONRawBody(c), http.StatusOK, nil
}
func GetJSONRawBody(c echo.Context) map[string]interface{} {

	request := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&request)
	if err != nil {
		return nil
	}

	return request
}

func ParseUserID(uid string) int64 {
	id, _ := strconv.ParseInt(uid, 10, 64)
	return id
}