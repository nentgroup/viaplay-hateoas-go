package hal

import (
	"encoding/json"
	"testing"
)

type DummyStruct struct {
	Name string `json:"name"`
}

// dummyResourceJSON is the canonical marshaled form of NewResource(DummyStruct{"Dummy"}, "uri").
// It is shared across multiple tests to satisfy the goconst linter.
const dummyResourceJSON = `{"links":{"self":{"href":"uri"}},"name":"Dummy"}`

// dummyResourceWithFooLinkCollectionJSON is the marshaled form of a Dummy resource
// that has a "foo" link relation containing two links ("bar" and "bar2").
const dummyResourceWithFooLinkCollectionJSON = `{"links":{"foo":[{"href":"bar"},{"href":"bar2"}],"self":{"href":"uri"}},"name":"Dummy"}`

// dummyResourceWithTwoEmbeddedJSON is the marshaled form of a Dummy resource
// that embeds two child Dummy resources under the "foo" relation.
const dummyResourceWithTwoEmbeddedJSON = `{"embedded":{"foo":[{"links":{"self":{"href":"uri2"}},"name":"DummyEmbed"},{"links":{"self":{"href":"uri3"}},"name":"DummyEmbed2"}]},"links":{"self":{"href":"uri"}},"name":"Dummy"}`

func TestNewResource(t *testing.T) {
	ds := DummyStruct{"Dummy"}

	r := NewResource(ds, "uri")

	if len(r.Links) < 1 {
		t.Errorf("No links added to the new resource")
	}

	if r.Links["self"] == nil {
		t.Errorf("No SELF link added to the new resource")
	}

	if len(r.Embedded) > 0 {
		t.Errorf("Embedded list should be empty")
	}
}

func TestLinkMarshal(t *testing.T) {
	l := make(Link)
	l["href"] = "http://localhost/"

	jl, err := json.Marshal(l)
	if err != nil {
		t.Errorf("%s", err)
	}

	if string(jl) != `{"href":"http://localhost/"}` {
		t.Errorf("Wrong Link struct: %s", jl)
	}
}

func TestResourceMarshal(t *testing.T) {
	expected := dummyResourceJSON

	ds := DummyStruct{"Dummy"}

	r := NewResource(ds, "uri")

	jr, err := json.Marshal(r)
	if err != nil {
		t.Errorf("%s", err)
	}

	if string(jr) != expected {
		t.Errorf("Wrong Resource struct: %v\n- Given: %v\n- Expected: %s", r, jr, expected)
	}
}

type DummyStructWithMapper struct {
	Name string
}

func (dswm DummyStructWithMapper) GetMap() Entry {
	return Entry{
		"customName": dswm.Name,
	}
}

func TestResourceMarshallWithMapper(t *testing.T) {
	expected := `{"customName":"Dummy","links":{"self":{"href":"uri"}}}`

	ds := DummyStructWithMapper{"Dummy"}

	r := NewResource(ds, "uri")

	jr, err := json.Marshal(r)
	if err != nil {
		t.Errorf("%s", err)
	}

	if string(jr) != expected {
		t.Errorf("Wrong Resource struct: %v\n- Given: %v\n- Expected: %s", r, jr, expected)
	}
}

func TestResourceMarshalWithoutLinks(t *testing.T) {
	expected := `{"name":"Dummy"}`

	ds := DummyStruct{"Dummy"}
	r := NewResource(ds, "")

	jr, err := json.Marshal(r)
	if err != nil {
		t.Errorf("%s", err)
	}

	if string(jr) != expected {
		t.Errorf("Wrong Resource struct: %v\n- Given: %v\n- Expected: %s", r, jr, expected)
	}
}

/* Test Links */
func TestNewLink(t *testing.T) {
	expected := `{"href":"bar","templated":true}`

	l := NewLink("bar", LinkAttr{"templated": true})

	jr, err := json.Marshal(l)
	if err != nil {
		t.Errorf("%s", err)
	}

	if string(jr) != expected {
		t.Errorf("Wrong link struct: %s\n- Given: %s\n- Expected: %s", l, jr, expected)
	}
}

func TestNewLinkMultipleAttributes(t *testing.T) {
	expected := `{"href":"http://haltalk.herokuapp.com/docs/{rel}","name":"doc","templated":true}`

	l := NewLink("http://haltalk.herokuapp.com/docs/{rel}", LinkAttr{"name": "doc"}, LinkAttr{"templated": true})

	jr, err := json.Marshal(l)
	if err != nil {
		t.Errorf("%s", err)
	}

	if string(jr) != expected {
		t.Errorf("Wrong link struct: %s\n- Given: %s\n- Expected: %s", l, jr, expected)
	}
}

func TestRegisterCurie(t *testing.T) {
	expected := `{"links":{"curies":[{"href":"http://haltalk.herokuapp.com/docs/{rel}","name":"doc","templated":true}],"doc:foo":{"href":"bar"},"self":{"href":"uri"}},"name":"Dummy"}`

	ds := DummyStruct{"Dummy"}

	r := NewResource(ds, "uri")
	r.RegisterCurie("doc", "http://haltalk.herokuapp.com/docs/{rel}", true).AddNewLink("foo", "bar")

	jr, err := json.Marshal(r)
	if err != nil {
		t.Errorf("%s", err)
	}

	if string(jr) != expected {
		t.Errorf("Wrong Resource struct: %v\n- Given: %v\n- Expected: %s", r, jr, expected)
	}
}

func TestRegisterMultipleCuries(t *testing.T) {
	expected := `{"links":{"curies":[{"href":"http://haltalk.herokuapp.com/docs/{rel}","name":"doc","templated":true},{"href":"http://haltalk.herokuapp.com/abc/{rel}","name":"abc","templated":true}],"doc:foo":{"href":"bar"},"self":{"href":"uri"}},"name":"Dummy"}`

	ds := DummyStruct{"Dummy"}

	r := NewResource(ds, "uri")
	r.RegisterCurie("doc", "http://haltalk.herokuapp.com/docs/{rel}", true).AddNewLink("foo", "bar")
	r.RegisterCurie("abc", "http://haltalk.herokuapp.com/abc/{rel}", true)

	jr, err := json.Marshal(r)
	if err != nil {
		t.Errorf("%s", err)
	}

	if string(jr) != expected {
		t.Errorf("Wrong Resource struct: %v\n- Given: %v\n- Expected: %s", r, jr, expected)
	}
}

func TestResourceCuries(t *testing.T) {
	ds := DummyStruct{"Dummy"}
	curieName := "doc"

	r := NewResource(ds, "uri")
	curie := r.RegisterCurie(curieName, "http://haltalk.herokuapp.com/docs/{rel}", true)

	curie.AddNewLink("foo", "bar")
	curies := r.Curies

	if len(curies) != 1 {
		t.Errorf("Wrong number of CurieHandles returned from resource:\n - Given: %v\n- Expected: %v\n", len(curies), 1)
	}

	if curies[curieName].Resource != r {
		t.Errorf("CurieHandle.Resource does not reference owning resource")
	}

	if curie != curies[curieName] {
		t.Errorf("curieHandle returned by RegisterCurie() is not the same reference")
	}
}

func TestAddNewLink(t *testing.T) {
	expected := `{"links":{"foo":{"href":"bar"},"self":{"href":"uri"}},"name":"Dummy"}`

	ds := DummyStruct{"Dummy"}

	r := NewResource(ds, "uri")
	r.AddNewLink("foo", "bar")

	jr, err := json.Marshal(r)
	if err != nil {
		t.Errorf("%s", err)
	}

	if string(jr) != expected {
		t.Errorf("Wrong Resource struct: %v\n- Given: %v\n- Expected: %s", r, jr, expected)
	}
}

func TestAddNewLinkTwice(t *testing.T) {
	expected := dummyResourceWithFooLinkCollectionJSON

	ds := DummyStruct{"Dummy"}

	r := NewResource(ds, "uri")
	r.AddNewLink("foo", "bar")
	r.AddNewLink("foo", "bar2")

	jr, err := json.Marshal(r)
	if err != nil {
		t.Errorf("%s", err)
	}

	if string(jr) != expected {
		t.Errorf("Wrong Resource struct: %v\n- Given:    %v\n- Expected: %s", r, jr, expected)
	}
}

func TestAddLinkCollection(t *testing.T) {
	expected := dummyResourceWithFooLinkCollectionJSON

	ds := DummyStruct{"Dummy"}

	r := NewResource(ds, "uri")
	r.AddLinkCollection("foo", LinkCollection{NewLink("bar", nil), NewLink("bar2", nil)})

	jr, err := json.Marshal(r)
	if err != nil {
		t.Errorf("%s", err)
	}

	if string(jr) != expected {
		t.Errorf("Wrong Resource struct: %v\n- Given:    %v\n- Expected: %s", r, jr, expected)
	}
}

func TestAddLinkCollectionToLink(t *testing.T) {
	expected := `{"links":{"foo":[{"href":"baz"},{"href":"bar"},{"href":"bar2"}],"self":{"href":"uri"}},"name":"Dummy"}`

	ds := DummyStruct{"Dummy"}

	r := NewResource(ds, "uri")
	r.AddNewLink("foo", "baz")
	r.AddLinkCollection("foo", LinkCollection{NewLink("bar", nil), NewLink("bar2", nil)})

	jr, err := json.Marshal(r)
	if err != nil {
		t.Errorf("%s", err)
	}

	if string(jr) != expected {
		t.Errorf("Wrong Resource struct: %v\n- Given:    %v\n- Expected: %s", r, jr, expected)
	}
}

/* Test Embedded */
func TestEmbed(t *testing.T) {
	expected := `{"embedded":{"foo":[{"links":{"self":{"href":"uri2"}},"name":"DummyEmbed"}]},"links":{"self":{"href":"uri"}},"name":"Dummy"}`

	ds := DummyStruct{"Dummy"}
	ds2 := DummyStruct{"DummyEmbed"}

	r := NewResource(ds, "uri")
	r2 := NewResource(ds2, "uri2")
	r.Embed("foo", r2)

	jr, err := json.Marshal(r)
	if err != nil {
		t.Errorf("%s", err)
	}

	if string(jr) != expected {
		t.Errorf("Wrong Resource struct: %v\n- Given: %v\n- Expected: %s", r, string(jr), expected)
	}
}

func TestEmbedTwice(t *testing.T) {
	expected := dummyResourceWithTwoEmbeddedJSON

	ds := DummyStruct{"Dummy"}
	ds2 := DummyStruct{"DummyEmbed"}
	ds3 := DummyStruct{"DummyEmbed2"}

	r := NewResource(ds, "uri")
	r2 := NewResource(ds2, "uri2")
	r3 := NewResource(ds3, "uri3")
	r.Embed("foo", r2)
	r.Embed("foo", r3)

	jr, err := json.Marshal(r)
	if err != nil {
		t.Errorf("%s", err)
	}

	if string(jr) != expected {
		t.Errorf("Wrong Resource struct: %v\n- Given:    %v\n- Expected: %s", r, jr, expected)
	}
}

func TestAddResourceCollection(t *testing.T) {
	expected := dummyResourceWithTwoEmbeddedJSON

	ds := DummyStruct{"Dummy"}
	ds2 := DummyStruct{"DummyEmbed"}
	ds3 := DummyStruct{"DummyEmbed2"}

	r := NewResource(ds, "uri")
	r2 := NewResource(ds2, "uri2")
	r3 := NewResource(ds3, "uri3")
	r.EmbedCollection("foo", ResourceCollection{r2, r3})

	jr, err := json.Marshal(r)
	if err != nil {
		t.Errorf("%s", err)
	}

	if string(jr) != expected {
		t.Errorf("Wrong Resource struct: %v\n- Given:    %v\n- Expected: %s", r, jr, expected)
	}
}

func TestAddResourceCollectionToResource(t *testing.T) {
	expected := `{"embedded":{"foo":[{"links":{"self":{"href":"uri2"}},"name":"DummyEmbed"},{"links":{"self":{"href":"uri3"}},"name":"DummyEmbed2"},{"links":{"self":{"href":"uri4"}},"name":"DummyEmbed3"}]},"links":{"self":{"href":"uri"}},"name":"Dummy"}`

	ds := DummyStruct{"Dummy"}
	ds2 := DummyStruct{"DummyEmbed"}
	ds3 := DummyStruct{"DummyEmbed2"}
	ds4 := DummyStruct{"DummyEmbed3"}

	r := NewResource(ds, "uri")
	r2 := NewResource(ds2, "uri2")
	r3 := NewResource(ds3, "uri3")
	r4 := NewResource(ds4, "uri4")
	r.Embed("foo", r2)
	r.EmbedCollection("foo", ResourceCollection{r3, r4})

	jr, err := json.Marshal(r)
	if err != nil {
		t.Errorf("%s", err)
	}

	if string(jr) != expected {
		t.Errorf("Wrong Resource struct: %v\n- Given:    %v\n- Expected: %s", r, jr, expected)
	}
}

func TestOmitEmptyReflection(t *testing.T) {
	expected := `{"id":null,"links":{"self":{"href":"test"}}}`
	dummyStruct := struct {
		ID *int `json:"id,omitempty"`
	}{}
	r := NewResource(dummyStruct, "test")
	res, err := json.Marshal(r)
	if err != nil {
		t.Errorf("%s", err)
	}
	if string(res) != expected {
		t.Errorf("Wrong Resource struct: %v\n- Given:    %v\n- Expected: %s", r, res, expected)
	}
}

func TestAddLinkOnResourceWithoutSelf(t *testing.T) {
	// Regression: NewResource with empty selfUri leaves Links nil; AddNewLink must
	// lazily initialize the map instead of panicking.
	expected := `{"links":{"foo":{"href":"bar"}},"name":"Dummy"}`

	r := NewResource(DummyStruct{"Dummy"}, "")
	r.AddNewLink("foo", "bar")

	jr, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}

func TestAddLinkCollectionOnResourceWithoutSelf(t *testing.T) {
	expected := `{"links":{"foo":[{"href":"bar"}]},"name":"Dummy"}`

	r := NewResource(DummyStruct{"Dummy"}, "")
	r.AddLinkCollection("foo", LinkCollection{NewLink("bar", nil)})

	jr, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}

func TestPointerPayload(t *testing.T) {
	// Regression: getPayloadMap must dereference pointer payloads instead of panicking.
	expected := dummyResourceJSON

	ds := &DummyStruct{"Dummy"}
	r := NewResource(ds, "uri")

	jr, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}

func TestNilPointerPayload(t *testing.T) {
	// Regression: a nil pointer payload should not panic; payload contributes no fields.
	expected := `{"links":{"self":{"href":"uri"}}}`

	var ds *DummyStruct
	r := NewResource(ds, "uri")

	jr, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}

func TestUnexportedFieldsSkipped(t *testing.T) {
	// Regression: unexported fields would panic when calling Value.Interface().
	type mixed struct {
		Name   string `json:"name"`
		secret string //nolint:unused
	}
	expected := dummyResourceJSON

	r := NewResource(mixed{Name: "Dummy", secret: "shh"}, "uri")
	jr, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}

/* Tests for payloads implementing the Mapper interface (GetMap) */

// mapperPayload implements Mapper on a value receiver.
type mapperPayload struct {
	First string
	Last  string
	Age   int
}

func (m mapperPayload) GetMap() Entry {
	return Entry{
		"fullName": m.First + " " + m.Last,
		"age":      m.Age,
	}
}

// pointerMapperPayload implements Mapper on a pointer receiver.
type pointerMapperPayload struct {
	Title string
}

func (p *pointerMapperPayload) GetMap() Entry {
	return Entry{
		"title": p.Title,
	}
}

// emptyMapperPayload returns no fields at all.
type emptyMapperPayload struct{}

func (emptyMapperPayload) GetMap() Entry { return Entry{} }

// nestedMapperPayload returns nested objects/arrays from GetMap.
type nestedMapperPayload struct {
	Name string
	Tags []string
}

func (n nestedMapperPayload) GetMap() Entry {
	return Entry{
		"name": n.Name,
		"meta": Entry{
			"tags":  n.Tags,
			"count": len(n.Tags),
		},
	}
}

// linksMapperPayload deliberately returns a "links" key from GetMap.
// The resource's own Links must take precedence (documents current behavior).
type linksMapperPayload struct{}

func (linksMapperPayload) GetMap() Entry {
	return Entry{
		"links":   "should-be-overwritten",
		"keptKey": "keptValue",
	}
}

func TestMapperPayloadMarshal(t *testing.T) {
	expected := `{"age":42,"fullName":"Ada Lovelace","links":{"self":{"href":"/people/1"}}}`

	r := NewResource(mapperPayload{First: "Ada", Last: "Lovelace", Age: 42}, "/people/1")

	jr, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}

func TestMapperPayloadWithExtraLink(t *testing.T) {
	expected := `{"age":42,"fullName":"Ada Lovelace","links":{"profile":{"href":"/profiles/1"},"self":{"href":"/people/1"}}}`

	r := NewResource(mapperPayload{First: "Ada", Last: "Lovelace", Age: 42}, "/people/1")
	r.AddNewLink("profile", "/profiles/1")

	jr, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}

func TestMapperPayloadWithEmbedded(t *testing.T) {
	expected := `{"age":42,"embedded":{"child":[{"age":7,"fullName":"Augusta King","links":{"self":{"href":"/people/2"}}}]},"fullName":"Ada Lovelace","links":{"self":{"href":"/people/1"}}}`

	parent := NewResource(mapperPayload{First: "Ada", Last: "Lovelace", Age: 42}, "/people/1")
	child := NewResource(mapperPayload{First: "Augusta", Last: "King", Age: 7}, "/people/2")
	parent.Embed("child", child)

	jr, err := json.Marshal(parent)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}

func TestPointerMapperPayload(t *testing.T) {
	// Mapper implemented on *T: passing the pointer satisfies the interface.
	expected := `{"links":{"self":{"href":"/posts/1"}},"title":"Hello"}`

	r := NewResource(&pointerMapperPayload{Title: "Hello"}, "/posts/1")

	jr, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}

func TestPointerMapperPayloadValueDoesNotImplementMapper(t *testing.T) {
	// When Mapper is on the pointer receiver and the user passes the value type,
	// the interface assertion fails and reflection-based payload mapping is used.
	// pointerMapperPayload has a single exported field "Title" with no json tag,
	// so it is emitted under its Go field name.
	expected := `{"Title":"Hello","links":{"self":{"href":"/posts/1"}}}`

	r := NewResource(pointerMapperPayload{Title: "Hello"}, "/posts/1")

	jr, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}

func TestMapperPayloadEmpty(t *testing.T) {
	expected := `{"links":{"self":{"href":"/x"}}}`

	r := NewResource(emptyMapperPayload{}, "/x")

	jr, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}

func TestMapperPayloadEmptyWithoutLinks(t *testing.T) {
	// Mapper returns nothing AND no self URI -> JSON should be the empty object.
	expected := `{}`

	r := NewResource(emptyMapperPayload{}, "")

	jr, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}

func TestMapperPayloadWithNestedEntries(t *testing.T) {
	expected := `{"links":{"self":{"href":"/items/1"}},"meta":{"count":2,"tags":["go","hal"]},"name":"Item"}`

	r := NewResource(nestedMapperPayload{Name: "Item", Tags: []string{"go", "hal"}}, "/items/1")

	jr, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}

func TestMapperPayloadLinksKeyIsOverwritten(t *testing.T) {
	// Documents current behavior: a "links" key returned by Mapper.GetMap is
	// overwritten by the resource's own Links during marshaling.
	expected := `{"keptKey":"keptValue","links":{"self":{"href":"/x"}}}`

	r := NewResource(linksMapperPayload{}, "/x")

	jr, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}

func TestMapperGetMapDirectlyOnResource(t *testing.T) {
	// Resource itself implements Mapper. Calling GetMap on a resource whose
	// payload is also a Mapper should merge the payload's entries with links/embedded.
	r := NewResource(mapperPayload{First: "Ada", Last: "Lovelace", Age: 42}, "/people/1")
	r.AddNewLink("profile", "/profiles/1")

	entry := r.GetMap()

	if entry["fullName"] != "Ada Lovelace" {
		t.Errorf("expected fullName from payload mapper, got %v", entry["fullName"])
	}
	if entry["age"] != 42 {
		t.Errorf("expected age=42 from payload mapper, got %v", entry["age"])
	}
	links, ok := entry["links"].(LinkRelations)
	if !ok {
		t.Fatalf("expected entry[\"links\"] to be LinkRelations, got %T", entry["links"])
	}
	if _, hasSelf := links["self"]; !hasSelf {
		t.Errorf("expected self link in entry[\"links\"]")
	}
	if _, hasProfile := links["profile"]; !hasProfile {
		t.Errorf("expected profile link in entry[\"links\"]")
	}
}
