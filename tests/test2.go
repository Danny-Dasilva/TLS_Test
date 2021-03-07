package main

import (
    "fmt"
    "io/ioutil"
    "net"
    "net/http"
    "net/http/cookiejar"
	"context"
    tls "github.com/refraction-networking/utls"
)

func main() {

    

    cookieJar, _ := cookiejar.New(nil)

    client := &http.Client{
		Jar: cookieJar,
		Transport: &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// Note that hardcoding the address is not necessary here. Only
				// do that if you want to ignore the DNS lookup that already
				// happened behind the scenes.
	
				tcpConn, err := (&net.Dialer{}).DialContext(ctx, network, addr) 
				if err != nil {
					return nil, err
				}
				config := tls.Config{ServerName: "ja3er.com"}
				tlsConn := tls.UClient(tcpConn, &config, tls.HelloChrome_Auto)
	
				err = tlsConn.Handshake()
				if err != nil {
					return nil, fmt.Errorf("uTlsConn.Handshake() error: %w", err)
				}
	
				return tlsConn, nil
			},
		},
	}

    req, err := http.NewRequest("GET", "https://ja3er.com/json", nil)
    req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.116 Safari/537.36")

    resp, err := client.Do(req)
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Println(resp.StatusCode)
        body, _ := ioutil.ReadAll(resp.Body)
        fmt.Println(string(body))
    }
}