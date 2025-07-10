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

	script, found := findScriptWithSubstrings(response, []string{"questions", "pack"})
	if !found {
		return nil, fmt.Errorf("failed to find <script> with questions")
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

// findScriptWithSubstrings searches for a script element whose content includes all the given substrings
func findScriptWithSubstrings(response *http.Response, substrings []string) (string, bool) {
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return "", false
	}

	var result string
	found := false

	doc.Find("script").EachWithBreak(func(i int, s *goquery.Selection) bool {
		scriptContent := strings.TrimSpace(s.Text())

		allPresent := true
		for _, substring := range substrings {
			if !strings.Contains(scriptContent, substring) {
				allPresent = false
				break
			}
		}

		if allPresent {
			result = scriptContent
			found = true
			return false
		}
		return true
	})

	return result, found
}

// FindKeyInData recursively searches for the key in nested data
func FindKeyInData(data any, key string) (any, error) {
	switch v := data.(type) {
	case map[string]any:
		if value, exists := v[key]; exists {
			return value, nil
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
