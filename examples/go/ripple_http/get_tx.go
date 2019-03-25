package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	resp, err := http.Get("https://testnet.data.api.ripple.com/v2/transactions/4C3AF3C9200289A0EA970CFE21F698DC6F3BBAEB3CB78E63CA3598A2F7FED5E9")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(body))
}
