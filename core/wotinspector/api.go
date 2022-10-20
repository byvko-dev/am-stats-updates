package wotinspector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/byvko-dev/am-core/helpers/env"
)

var apiURL = env.MustGetString("WOT_INSPECTOR_API_URI")

func GetWotInspectorTanks() (map[int]VehicleInfo, error) {
	re := regexp.MustCompile(`(\d{1,9}):`)
	tanks := make(map[int]VehicleInfo)

	res, err := http.DefaultClient.Get(apiURL)
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		return tanks, fmt.Errorf("status code: %+v. error: %s", res, err)
	}
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return tanks, fmt.Errorf("status code: %+v. error: %s", res, err)
	}
	tanksString := strings.ReplaceAll(string(bodyBytes), "TANK_DB = ", "")
	tanksString = re.ReplaceAllString(tanksString, `"$1":`)
	split := strings.SplitAfter(tanksString, "},")
	if len(split) <= 2 {
		return tanks, fmt.Errorf("failed to split string")
	}
	fix := strings.ReplaceAll(split[len(split)-2], "},", "}")
	tanksString = strings.ReplaceAll(tanksString, split[len(split)-2], fix)
	return tanks, json.Unmarshal([]byte(tanksString), &tanks)
}
