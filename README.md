<img src="./.github/assets/gopher.png" align="right" height="96" width="96"/>

<br />

# Viaplay HATEOAS for Go
[![Tests](https://github.com/nentgroup/viaplay-hateoas-go/actions/workflows/test.yml/badge.svg)](https://github.com/nentgroup/viaplay-hateoas-go/actions/workflows/test.yml)

A Go library for building hypermedia-driven RESTful APIs, following the [Viaplay HATEOAS standard](http://github.com/nentgroup/api-guidelines).

## <picture><source media="(prefers-color-scheme: dark)" srcset="./.github/assets/icons/settings-dark.svg"><img src="./.github/assets/icons/settings-light.svg" alt="Installation" width="20" height="20" aria-label="Installation"></picture> Installation

```bash
go get github.com/nentgroup/viaplay-hateoas-go/v2
```

---

## <picture><source media="(prefers-color-scheme: dark)" srcset="./.github/assets/icons/rocket-dark.svg"><img src="./.github/assets/icons/rocket-light.svg" alt="Usage" width="20" height="20" aria-label="Usage"></picture> Usage

Map your Go structs into HAL resources by implementing the `hal.Mapper` interface — just declare which fields to expose and how:

```go
type Mapper interface {
	GetMap() Entry
}
```

### <picture><source media="(prefers-color-scheme: dark)" srcset="./.github/assets/icons/search-dark.svg"><img src="./.github/assets/icons/search-light.svg" alt="Basic Example" width="18" height="18" aria-label="Basic Example"></picture> Basic Example

For a given Product struct, this would be the `hal.Mapper` implementation:

```go
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
```

Then you can create a HAL Resource for a Product by:

```go
p := Product{
	Code:  1,
	Name:  "Some Product",
	Price: 10,
}

pr := hal.NewResource(p, "http://rest.api/products/1")
```

When marshalled to JSON, this produces (using the default `FlavorViaplay` flavor):

```json
{
	"data": {
		"name": "Some product",
		"price": 10
	},
	"links": {
		"self": {"href": "http://rest.api/products/1"}
	}
}
```

---

## <picture><source media="(prefers-color-scheme: dark)" srcset="./.github/assets/icons/layers-dark.svg"><img src="./.github/assets/icons/layers-light.svg" alt="Embedded Resources" width="20" height="20" aria-label="Embedded Resources"></picture> Embedded Resources

In HAL, even the API response itself is a resource, and it can embed other resources. Say your API serves a list of Task structs:

```go
type (
	Response struct {
		Count int
		Total int
	}

	Task struct {
		Id   int
		Name string
	}
)

func (p Response) GetMap() hal.Entry {
	return hal.Entry{
		"count": p.Count,
		"total": p.Total,
	}
}

func (c Task) GetMap() hal.Entry {
	return hal.Entry{
		"id":   c.Id,
		"name": c.Name,
	}
}
```

Create the parent resource, then embed child resources under a relation name. Embed the same relation more than once and it becomes a collection; embed it once and it stays a single object:

```go
// Creating Response resource
r := hal.NewResource(Response{Count: 10, Total: 20}, "/tasks")
r.AddNewLink("next", "/tasks=page=2")

// Creating Task resources
t1 := hal.NewResource(Task{Id: 1, Name: "Some Task"}, "/tasks/1")
t2 := hal.NewResource(Task{Id: 2, Name: "Some Task"}, "/tasks/2")

// Embedding multiple resources as an array
r.Embed("tasks", t1)
r.Embed("tasks", t2)

// Or set a single embedded resource as an object
r.Embedded.Set("task", t1)
```

Use `Set` / `SetCollection` when you want to explicitly define a relation as a single object or a collection, and `Add` / `AddCollection` (or the `Embed` shortcut above) when you just want to keep appending to it.

This produces (using the default `FlavorViaplay` flavor):

```json
{
  "data": {
    "count": 10,
    "total": 20
  },
  "embedded": {
    "tasks": [
      {
        "data": {
          "id": 1,
          "name": "Some Task"
        },
        "links": {
          "self": {
            "href": "/tasks/1"
          }
        }
      },
      {
        "data": {
          "id": 2,
          "name": "Some Task"
        },
        "links": {
          "self": {
            "href": "/tasks/2"
          }
        }
      }
    ]
  },
  "links": {
    "next": {
      "href": "/tasks=page=2"
    },
    "self": {
      "href": "/tasks"
    }
  }
}
```

---

## <picture><source media="(prefers-color-scheme: dark)" srcset="./.github/assets/icons/link-dark.svg"><img src="./.github/assets/icons/link-light.svg" alt="CURIES" width="20" height="20" aria-label="CURIES"></picture> CURIES

To include CURIE relations in your output, register the curie name and fluently add a link relation:

```go
p := Product{
	Code:  1,
	Name:  "Some Product",
	Price: 10,
}

// Creating Product resource
pr := hal.NewResource(p, "http://rest.api/products/1")
pr.RegisterCurie("acme", "http://acme.com/docs/{rel}", true)
   .AddNewLink("widgets", "http://rest.api/products/1/widgets")
```

Output (using the default `FlavorViaplay` flavor):

```json
{
	"data": {
		"name": "Some product",
		"price": 10
	},
	"links": {
		"self": {"href": "http://rest.api/products/1"},
		"curies": [{ 
		        "name": "acme",
		        "href": "http://acme.com/docs/{rel}",
		        "templated": true
		    }],
		"acme:widgets": { "href": "http://rest.api/products/1/widgets" }
	}
}
```

Registered curies can also be retrieved by name from the resource's `Curies` map:

```go
pr := hal.NewResource(p, "http://rest.api/products/1")
pr.RegisterCurie("acme", "http://acme.com/docs/{rel}", true)
// ...

curie := pr.Curies["acme"]
curie.AddNewLink("widgets", "http://rest.api/products/1/widgets")
```

---

## <picture><source media="(prefers-color-scheme: dark)" srcset="./.github/assets/icons/settings-dark.svg"><img src="./.github/assets/icons/settings-light.svg" alt="Flavors" width="20" height="20" aria-label="Flavors"></picture> Flavors: Viaplay HAL vs canonical HAL

By default, every resource created with `hal.NewResource` uses the [Viaplay flavor](./specs/viaplay-hateoas.md): payload nested under a `data` key, with unprefixed `links`/`embedded` keys. You can opt a resource into the historical flattened shape (`FlavorDefault`) or into canonical [HAL](https://datatracker.ietf.org/doc/html/draft-kelly-json-hal) (`_links`/`_embedded`, `FlavorHAL`) by setting the `Flavor` field:

```go
pr := hal.NewResource(p, "/products/1")
pr.Flavor = hal.FlavorHAL // or hal.FlavorDefault
```

To switch an entire application at once, set the package-level default before creating any resources — every resource created afterwards will inherit it:

```go
hal.DefaultFlavor = hal.FlavorHAL
```

| Flavor            | Payload            | Keys                     |
|-------------------|---------------------|--------------------------|
| `FlavorViaplay` (default) | nested under `data` | `links` / `embedded`     |
| `FlavorHAL`        | flattened onto root | `_links` / `_embedded`   |
| `FlavorDefault`    | flattened onto root | `links` / `embedded`     |

Embedded resources are marshalled using their own `Flavor` field, so make sure child resources are created with (or assigned) the same flavor as their parent, e.g. by setting `hal.DefaultFlavor` once before building the resource graph.

### Side-by-side example

Given the same `Product` resource (with an embedded `category`), here is the JSON produced by each flavor:

**`FlavorViaplay`** (payload nested under `data`, unprefixed keys):

```json
{
  "data": {
    "name": "Some Product",
    "price": 10
  },
  "embedded": {
    "category": {
      "data": {
        "name": "Some Category",
        "price": 0
      },
      "links": {
        "self": { "href": "/categories/1" }
      }
    }
  },
  "links": {
    "self": { "href": "/products/1" }
  }
}
```

**`FlavorHAL`** (payload flattened, `_`-prefixed keys):

```json
{
  "_embedded": {
    "category": {
      "_links": {
        "self": { "href": "/categories/1" }
      },
      "name": "Some Category",
      "price": 0
    }
  },
  "_links": {
    "self": { "href": "/products/1" }
  },
  "name": "Some Product",
  "price": 10
}
```

---

## <picture><source media="(prefers-color-scheme: dark)" srcset="./.github/assets/icons/document-dark.svg"><img src="./.github/assets/icons/document-light.svg" alt="XML Output" width="20" height="20" aria-label="XML Output"></picture> XML Output

`Resource` implements `xml.Marshaler`, so the same resource graph you build for JSON can also be marshalled to a HAL+XML representation. XML output is flavor-agnostic — payload fields are always flattened as child elements, regardless of the resource's `Flavor`:

```go
pr := hal.NewResource(p, "/products/1")
pr.Embed("category", hal.NewResource(c, "/categories/1"))

x, err := xml.MarshalIndent(pr, "", "  ")
```

Output:

```xml
<resource href="/products/1">
  <name>Some Product</name>
  <price>10</price>
  <resource rel="category" href="/categories/1">
    <name>Some Category</name>
  </resource>
</resource>
```

See `examples/xml` for a full example.

---

## <picture><source media="(prefers-color-scheme: dark)" srcset="./.github/assets/icons/chart-dark.svg"><img src="./.github/assets/icons/chart-light.svg" alt="Generics Helpers" width="20" height="20" aria-label="Generics Helpers"></picture> Generics Helpers

A few small generic helpers reduce boilerplate around typed payloads:

```go
// Typed payload retrieval
product, ok := hal.Payload[Product](pr)

// Adapt a type you don't own to hal.Mapper without modifying it
mapped := hal.WithMapper(externalDTO, func(d ExternalDTO) hal.Entry {
	return hal.Entry{"name": d.Name}
})

// Build a ResourceCollection from a typed slice without a manual loop
resources := hal.NewResourceCollection(products, func(p Product) string {
	return fmt.Sprintf("/products/%d", p.Code)
})
```

See `examples/generics` for a full example.

---

## <picture><source media="(prefers-color-scheme: dark)" srcset="./.github/assets/icons/rocket-dark.svg"><img src="./.github/assets/icons/rocket-light.svg" alt="Migrating from v1" width="20" height="20" aria-label="Migrating from v1"></picture> Migrating from v1

Upgrading from v1? See [MIGRATION.md](./docs/MIGRATION.md) for the full list of breaking changes and how to adapt your code.

---

## License

This project is licensed under the MIT License - see the LICENSE file for details.
