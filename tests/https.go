package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	// "net/url"
	// "path"
	"github.com/CUCyber/ja3transport"
)

// JA3Response is the struct
type JA3Response struct {
	JA3Hash   string `json:"ja3_hash"`
	JA3       string `json:"ja3"`
	UserAgent string `json:"User-Agent"`
}

type test_struct struct {
    Test string
}


func main() {
	site := "https://clienttest.ssllabs.com:8443/ssltest/viewMyClient.html"
	
	httpClient,err := ja3transport.New(ja3transport.SafariAuto)
	// httpClient, err := ja3transport.NewWithString("771,4865-4866-4867-49196-49195-49188-49187-49162-49161-52393-49200-49199-49192-49191-49172-49171-52392-157-156-61-60-53-47-49160-49170-10,65281-0-23-13-5-18-16-11-51-45-43-10-21,29-23-24-25,0")
	if err != nil{
		fmt.Println(err)
		panic(err)
	}

	/* First fetch the JA3 Fingerprint */
	resp, err := httpClient.Get(site)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	// unmarshal the response
	var ja3Response JA3Response
	err = json.Unmarshal(body, &ja3Response)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// /* Fetch information about the ja3hash*/
	// searchURL, _ := url.Parse("https://gobyexample.com/variables")
	// searchURL.Path = path.Join(searchURL.Path, ja3Response.JA3Hash)

	// resp, err = httpClient.Get(searchURL.String())
	// if err != nil {
	// 	fmt.Println(err)
	// 	panic(err)
	// }

	// body, err = ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	fmt.Println(err)
	// 	panic(err)
	// }

	var out bytes.Buffer
	err = json.Indent(&out, body, "", "\t")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	fmt.Println(out.String())
}