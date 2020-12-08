package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
)

func main() {
	proxyUrl, _ := url.Parse("http://localhost:8080")
	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
	resp, _ := client.Get("http://ip.gs")
	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	print(string(bodyBytes))
}
