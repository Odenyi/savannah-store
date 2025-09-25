package library
import (
	"github.com/labstack/echo/v4"
	"net/http"
	"encoding/json"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"

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



type TokenResponse struct {
	Token string `json:"token"`
	// Add other fields if the API returns more
}

// GetIntouchToken generates a token for Intouch SMS API
func GetSMSToken() (string, error) {
	msisdn := os.Getenv("VL_MSISDN")
	password := os.Getenv("VL_PASSWORD")
	url := os.Getenv("VL_TOKENURL")

	if msisdn == "" || password == "" {
		return "", fmt.Errorf(" credentials are missing")
	}

	// Encode credentials
	credentials := fmt.Sprintf("%s:%s", msisdn, password)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", encoded))
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get token, status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}

	return tokenResp.Token, nil
}
