package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func GetJSON(url string, target interface{}) error {
	res, err := http.DefaultClient.Get(url)
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %+v. error: %s", res, err)
	}
	defer res.Body.Close()
	return json.NewDecoder(res.Body).Decode(target)
}
