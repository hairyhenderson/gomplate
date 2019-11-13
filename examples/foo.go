package main

import (
	"fmt"

	"github.com/hairyhenderson/gomplate/v3/data"
)

func main() {
	d, err := data.NewData([]string{"ip=https://ipinfo.io"}, nil)
	if err != nil {
		panic(err)
	}

	response, err := d.Datasource("ip")
	if err != nil {
		panic(err)
	}

	m := response.(map[string]interface{})
	fmt.Printf("country is %s\n", m["country"])
}
