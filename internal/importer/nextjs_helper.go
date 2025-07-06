package importer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ExtractNextJsDataFromURL fetches a URL and extracts Next.js data
func ExtractNextJsDataFromURL(url string) (any, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL %s: %w", url, err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d when fetching %s", response.StatusCode, url)
	}

	defer response.Body.Close()

	script, found := findScriptWithSubstring(response, "packs")
	if !found {
		return nil, fmt.Errorf("failed to find <script> packs")
	}

	fixedScript, err := extractJSONFromNextJSPush(script)
	if err != nil {
		return nil, fmt.Errorf("failed to extract JSON from push(): %w", err)
	}

	var parsedJSON any
	err = json.Unmarshal([]byte(fixedScript), &parsedJSON)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	return parsedJSON, nil
}

// extractJSONFromNextJSPush extracts the JSON string from a self.__next_f.push() call
func extractJSONFromNextJSPush(line string) (string, error) {
	line = strings.TrimSpace(line)

	// Regex to match self.__next_f.push([number, "almost_json_string"])
	re := regexp.MustCompile(`self\.__next_f\.push\(\[.*?,"(.*)"\]\)`)
	matches := re.FindStringSubmatch(line)

	if len(matches) < 2 {
		return "", fmt.Errorf("could not find JSON string in the line")
	}

	jsonStr := matches[1]
	// almost_json_string starts with "5:" and ends with \n
	jsonStr = jsonStr[2 : len(jsonStr)-2]
	jsonStr = strings.ReplaceAll(jsonStr, `\"`, `"`)
	jsonStr = strings.ReplaceAll(jsonStr, `\\`, `\`)

	return jsonStr, nil
}

// findScriptWithSubstring searches for a script element whose content starts with the given prefix
func findScriptWithSubstring(response *http.Response, substring string) (string, bool) {
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return "", false
	}

	var result string
	found := false

	doc.Find("script").EachWithBreak(func(i int, s *goquery.Selection) bool {
		scriptContent := strings.TrimSpace(s.Text())

		if strings.Contains(scriptContent, substring) {
			result = scriptContent
			found = true
			return false
		}
		return true
	})

	return result, found
}

func ExtractNextPropsFromURL(url string) (map[string]any, error) {
	var propsData map[string]any
	return propsData, nil
}

// FindKeyInData recursively searches for the key in nested data
func FindKeyInData(data any, key string) ([]any, error) {
	switch v := data.(type) {
	case map[string]any:
		if value, exists := v[key]; exists {
			if mapValue, ok := value.([]any); ok {
				return mapValue, nil
			}
			return nil, fmt.Errorf("key '%s' found but value is not a list", key)
		}

		for _, value := range v {
			if result, err := FindKeyInData(value, key); err == nil {
				return result, nil
			}
		}

	case []any:
		for _, item := range v {
			if result, err := FindKeyInData(item, key); err == nil {
				return result, nil
			}
		}
	}

	return nil, fmt.Errorf("key '%s' not found in data", key)
}
