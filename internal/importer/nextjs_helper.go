package importer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ExtractNextProps extracts Next.js props from a HTTP response
func ExtractNextProps(response *http.Response) (map[string]interface{}, error) {
	defer response.Body.Close()
	
	// Parse HTML content
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	
	// Find the __NEXT_DATA__ script tag
	var nextDataScript string
	doc.Find("script#__NEXT_DATA__").Each(func(i int, s *goquery.Selection) {
		nextDataScript = s.Text()
	})
	
	if nextDataScript == "" {
		return nil, fmt.Errorf("could not find Next.js data in the HTML")
	}
	
	// Parse JSON data
	var propsData map[string]interface{}
	err = json.Unmarshal([]byte(strings.TrimSpace(nextDataScript)), &propsData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON data from script tag: %w", err)
	}
	
	return propsData, nil
}

// ExtractNextPropsFromURL fetches a URL and extracts Next.js props
func ExtractNextPropsFromURL(url string) (map[string]interface{}, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL %s: %w", url, err)
	}
	
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d when fetching %s", response.StatusCode, url)
	}
	
	return ExtractNextProps(response)
}