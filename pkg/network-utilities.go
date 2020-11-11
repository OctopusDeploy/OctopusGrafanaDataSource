package main

import (
	"io/ioutil"
	"net/http"
)

func httpGet(url string) (result []byte, err error) {
	resp, err := http.Get(url)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
