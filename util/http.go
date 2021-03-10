package util

import (
	"io/ioutil"
	"net/http"
)

func HttpGet(url string) ([]byte, http.Header, error) {
	resp, err := http.Get(url)
	defer func() {
		if resp != nil {
			_ = resp.Body.Close()
		}
	}()
	if err != nil {
		return nil, nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return body, resp.Header, err
}
