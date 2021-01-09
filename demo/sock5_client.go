package demo

/* package main
 *
 * import (
 *     "io/ioutil"
 *     "net/http"
 *
 *     "golang.org/x/net/proxy"
 * )
 *
 * func main() {
 *     dialer, _ := proxy.SOCKS5("tcp", "127.0.0.1:8080", nil, proxy.Direct)
 *     client := &http.Client{
 *         Transport: &http.Transport{Dial: dialer.Dial},
 *     }
 *     resp, _ := client.Get("http://ip.gs")
 *     defer resp.Body.Close()
 *     bodyBytes, _ := ioutil.ReadAll(resp.Body)
 *     println(string(bodyBytes))
 * } */
