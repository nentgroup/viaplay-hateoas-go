package main

import (
	"encoding/json"
	"fmt"

	hal "github.com/nentgroup/viaplay-hateoas-go/v2"
)

// externalDTO simulates a struct defined in a package you don't own, so you
// cannot add a GetMap method to it directly.
type externalDTO struct {
	ID   int
	Name string
}

func main() {
	items := []externalDTO{
		{ID: 1, Name: "Widget"},
		{ID: 2, Name: "Gadget"},
	}

	// NewResourceCollection builds a ResourceCollection from typed items,
	// computing each resource's self link, without a manual loop.
	resources := hal.NewResourceCollection(items, func(d externalDTO) string {
		return fmt.Sprintf("/items/%d", d.ID)
	})

	// WithMapper lets you adapt a type to hal.Mapper without owning it or
	// writing a GetMap method for it, by attaching it as each resource's
	// payload afterwards.
	for i, item := range items {
		resources[i].Payload = hal.WithMapper(item, func(d externalDTO) hal.Entry {
			return hal.Entry{"name": d.Name}
		})
	}

	r := hal.NewResource(struct{}{}, "/items")
	r.EmbedCollection("items", resources)

	// Payload retrieves a typed payload back out of a Resource.
	if first, ok := hal.Payload[externalDTO](hal.NewResource(items[0], "")); ok {
		fmt.Printf("first item struct: %+v\n", first)
	}

	j, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	fmt.Printf("%s\n", j)
}
