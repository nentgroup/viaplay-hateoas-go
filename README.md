<img src="./.github/assets/gopher.png" align="right" height="96" width="96"/>


# Viaplay HATEOAS for Go
[![Tests](https://github.com/nentgroup/viaplay-hateoas-go/actions/workflows/tests.yml/badge.svg)](https://github.com/nentgroup/viaplay-hateoas-go/actions/workflows/tests.yml)

A Go implementation of the [Viaplay HATEOAS standard](http://github.com/nentgroup/api-guidelines) for creating hypermedia-driven RESTful APIs.


## 📦 Installation

```bash
go get github.com/nentgroup/viaplay-hateoas-go
```

---

## 🚀 Usage

This library allows you to map Go structs into HAL Resources by implementing the `hal.Mapper` interface. You only need to define which fields you want and how they should be represented.

```go
type Mapper interface {
	GetMap() Entry
}
```

### 🔍 Basic Example

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

When marshalled to JSON, this produces:

```json
{
	"_links": {
		"self": {"href": "http://rest.api/products/1"}
	},
	"name": "Some product",
	"price": 10
}
```

---

## 🔄 Embedded Resources

Let's say your API needs to serve a list of Task structs.

Since in HAL standard everything is a resource, even the API response itself can be treated as a resource containing other embedded resources:

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

### Creating and Embedding Resources

```go
// Creating Response resource
r := hal.NewResource(Response{Count: 10, Total: 20}, "/tasks")
r.AddNewLink("next", "/tasks=page=2")

// Creating Task resources
t1 := hal.NewResource(Task{Id: 1, Name: "Some Task"}, "/tasks/1")
t2 := hal.NewResource(Task{Id: 2, Name: "Some Task"}, "/tasks/2")

// Embedding
r.Embed("tasks", t1)
r.Embed("tasks", t2)
```

This produces:

```json
{
  "_embedded": {
    "tasks": [
      {
        "_links": {
          "self": {
            "href": "/tasks/1"
          }
        },
        "id": 1,
        "name": "Some Task"
      },
      {
        "_links": {
          "self": {
            "href": "/tasks/2"
          }
        },
        "id": 2,
        "name": "Some Task"
      }
    ]
  },
  "_links": {
    "next": {
      "href": "/tasks=page=2"
    },
    "self": {
      "href": "/tasks"
    }
  },
  "count": 10,
  "total": 20
}
```

---

## 🔗 CURIES

To include CURIE relations in your output, you can 'register' the curie name and fluently add a link relation:

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

Output:

```json
{
	"_links": {
		"self": {"href": "http://rest.api/products/1"},
		"curies": [{ 
		        "name": "acme",
		        "href": "http://acme.com/docs/{rel}",
		        "templated": true
		    }],
		"acme:widgets": { "href": "http://rest.api/products/1/widgets" }
	},
	"name": "Some product",
	"price": 10
}
```

### Alternative Method

Registered curies can also be retrieved by name from the resources' Curies map:

```go
pr := hal.NewResource(p, "http://rest.api/products/1")
pr.RegisterCurie("acme", "http://acme.com/docs/{rel}", true)
// ...

curie := pr.Curies["acme"]
curie.AddNewLink("widgets", "http://rest.api/products/1/widgets")
```

---

## 🛣️ Future Plans

- XML Marshaler support
- Improved documentation
- More examples

---

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.
