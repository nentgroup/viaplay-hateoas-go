package hal

import (
	"encoding/json"
	"strconv"
	"testing"
)

// benchPayload is a typical reflection-based payload (no Mapper).
type benchPayload struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Bio   string `json:"bio"`
	Age   int    `json:"age"`
}

// benchMapperPayload is a typical Mapper-based payload.
type benchMapperPayload struct {
	ID    int
	Name  string
	Email string
}

func (m benchMapperPayload) GetMap() Entry {
	return Entry{
		"id":    m.ID,
		"name":  m.Name,
		"email": m.Email,
	}
}

func newBenchResource() *Resource {
	r := NewResource(benchPayload{
		ID:    42,
		Name:  "Ada Lovelace",
		Email: "ada@example.com",
		Bio:   "Mathematician and writer.",
		Age:   36,
	}, "/people/42")
	r.AddNewLink("profile", "/profiles/42")
	r.AddNewLink("avatar", "/avatars/42.png")
	return r
}

func newBenchResourceWithEmbedded(n int) *Resource {
	parent := newBenchResource()
	children := make(ResourceCollection, 0, n)
	for i := 0; i < n; i++ {
		c := NewResource(benchPayload{
			ID:    i,
			Name:  "child-" + strconv.Itoa(i),
			Email: "c" + strconv.Itoa(i) + "@example.com",
			Bio:   "child bio",
			Age:   i,
		}, "/children/"+strconv.Itoa(i))
		c.AddNewLink("parent", "/people/42")
		children = append(children, c)
	}
	parent.EmbedCollection("children", children)
	return parent
}

func BenchmarkMarshal_SimpleResource(b *testing.B) {
	r := newBenchResource()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(r); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_MapperPayload(b *testing.B) {
	r := NewResource(benchMapperPayload{ID: 1, Name: "Ada", Email: "ada@example.com"}, "/people/1")
	r.AddNewLink("profile", "/profiles/1")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(r); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_Embedded10(b *testing.B) {
	r := newBenchResourceWithEmbedded(10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(r); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_Embedded100(b *testing.B) {
	r := newBenchResourceWithEmbedded(100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(r); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetPayloadMap_Reflection(b *testing.B) {
	r := &Resource{Payload: benchPayload{ID: 1, Name: "x", Email: "y", Bio: "z", Age: 9}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.getPayloadMap()
	}
}

func BenchmarkNewResource(b *testing.B) {
	p := benchPayload{ID: 1, Name: "x", Email: "y", Bio: "z", Age: 9}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewResource(p, "/x/1")
	}
}
