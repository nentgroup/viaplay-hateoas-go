package main

import (
	"encoding/xml"
	"fmt"

	hal "github.com/nentgroup/viaplay-hateoas-go/v2"
)

type Product struct {
	Code  int
	Name  string
	Price int
}

func (p Product) GetMap() hal.Entry {
	return hal.Entry{
		"name":  p.Name,
		"price": p.Price,
	}
}

func main() {
	p := Product{
		Code:  1,
		Name:  "Some Product",
		Price: 10,
	}

	category := hal.NewResource(Product{Name: "Some Category"}, "/categories/1")

	// Creating HAL Resource
	pr := hal.NewResource(p, "/products/1")
	pr.AddNewLink("help", "/docs")
	pr.Embed("category", category)

	// XML Encoding: the same Resource graph used for JSON can be
	// marshaled to HAL+XML by implementing xml.Marshaler.
	x, err := xml.MarshalIndent(pr, "", "  ")
	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	fmt.Printf("%s\n", x)
}
