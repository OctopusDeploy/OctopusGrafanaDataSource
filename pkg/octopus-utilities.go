package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

func slugify(value string) string {
	value = strings.ToLower(value)
	value = regexp.MustCompile(`\s`).ReplaceAllString(value, "-")
	value = regexp.MustCompile(`[^a-zA-Z0-9-]`).ReplaceAllString(value, "-")
	value = regexp.MustCompile(`-+`).ReplaceAllString(value, "-")
	value = strings.Trim(value, "-/")
	return value
}

func resourceNameToId(resourceType string, path string, space string, apiKey string, resourceName string) (string, error) {
	url := path + "/api/" + space + "/" + resourceType + "/" + slugify(resourceName) + "?apikey=" + apiKey
	resp, err := http.Get(url)
	defer resp.Body.Close()

	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	parsedResults := IdResource{}
	err = json.Unmarshal(body, &parsedResults)

	if err == nil {
		return parsedResults.Id, nil
	}

	return "", err
}
