// Copyright 2014 Nicolas Vellon.  All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package hal implements encoding of structs into HAL as defined in
// http://stateless.co/hal_specification.html.
//
// See the basic example for an introduction to this package:
// https://github.com/nvellon/hal/blob/master/examples/basic.go

package hal

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"
)

type (
	Entry map[string]any

	// Mapper is the interface implemented by the objects
	// that can be converted into HAL format.
	//
	// GetMap must return a freshly allocated Entry on every call (e.g. a
	// map literal), not a shared/cached map reused across calls: Resource's
	// marshaling may add "links"/"embedded" (or "data" nesting) entries
	// directly into the returned map for performance.
	Mapper interface {
		GetMap() Entry
	}

	Relation string

	// Flavor selects the JSON key names and payload layout used when
	// marshaling a Resource. See FlavorDefault, FlavorHAL and
	// FlavorViaplay for the supported flavors.
	Flavor int

	CurieHandle struct {
		Name string
		*Resource
	}

	// LinkAttr types that store hyperlinks and its attributes.
	LinkAttr       map[string]any
	Link           LinkAttr
	LinkCollection []Link
	LinkRelations  map[Relation]any

	// Resource is a struct that stores a resource data.
	// It represents a converted object in the HAL spec by
	// containing all its fields and also a set of related links
	// and a sub-set of recursively related resources.
	Resource struct {
		Payload  any
		Links    LinkRelations
		Embedded Embedded
		Curies   map[string]*CurieHandle
		// Flavor selects the key names and payload layout used when this
		// resource is marshaled to JSON. It defaults to DefaultFlavor
		// when created via NewResource, and can be overridden per-resource
		// by assigning the field directly (Resource has no constructor
		// options, so this is a plain exported field).
		Flavor Flavor
	}
	ResourceCollection []*Resource

	// Embedded stores related resources by relation. Internally the relation is
	// always represented as a collection so that the API is explicit and does not
	// mix object and array values in the same map.
	Embedded map[Relation]ResourceCollection
)

const (
	// FlavorDefault preserves this library's pre-v2 historical behavior:
	// payload fields are flattened directly onto the resource object, and
	// the "links"/"embedded" keys are unprefixed. It is the zero value of
	// Flavor, but is no longer the flavor NewResource assigns by default
	// (see DefaultFlavor); it remains available as an explicit opt-in for
	// code that depended on the old shape.
	FlavorDefault Flavor = iota

	// FlavorHAL produces canonical HAL documents as described by
	// https://datatracker.ietf.org/doc/html/draft-kelly-json-hal: payload
	// fields are flattened directly onto the resource object, and related
	// resources/links are exposed under the "_links" and "_embedded" keys.
	FlavorHAL

	// FlavorViaplay produces documents matching the flavor described in
	// specs/viaplay-hateoas.md: payload fields are nested under a "data"
	// key, and related resources/links are exposed under the unprefixed
	// "links" and "embedded" keys.
	FlavorViaplay
)

// DefaultFlavor is the Flavor assigned to resources created via NewResource.
// It defaults to FlavorViaplay since this package implements the Viaplay
// HATEOAS spec (specs/viaplay-hateoas.md) first and foremost; set it to
// FlavorHAL (or FlavorDefault for the legacy flattened/unprefixed shape)
// once during application start-up to switch the whole program instead of
// setting Resource.Flavor on every resource individually.
var DefaultFlavor = FlavorViaplay

// keys returns the JSON key names used for the links and embedded
// collections under this flavor.
func (f Flavor) keys() (linksKey, embeddedKey string) {
	if f == FlavorHAL {
		return "_links", "_embedded"
	}
	return "links", "embedded"
}

// AddNewLink adds a link to the resources_link collection
// prepended with the curie Name
func (c CurieHandle) AddNewLink(rel Relation, href string) {
	rel = Relation(c.Name) + ":" + rel
	c.AddLink(rel, NewLink(href, nil))
}

// AddCollection appends the resource collection into the list of embedded
// resources with the specified relation.
func (e Embedded) AddCollection(rel Relation, r ResourceCollection) {
	if len(r) == 0 || e == nil {
		return
	}
	e[rel] = append(append(ResourceCollection(nil), e[rel]...), r...)
}

// Add appends a resource to the relation.
func (e Embedded) Add(rel Relation, r *Resource) {
	if r == nil || e == nil {
		return
	}
	e[rel] = append(e[rel], r)
}

// Set replaces the relation with a single embedded resource.
func (e Embedded) Set(rel Relation, r *Resource) {
	if e == nil {
		return
	}
	if r == nil {
		delete(e, rel)
		return
	}
	e[rel] = ResourceCollection{r}
}

// SetCollection replaces the relation with an explicit collection of resources.
func (e Embedded) SetCollection(rel Relation, r ResourceCollection) {
	if e == nil {
		return
	}
	if len(r) == 0 {
		delete(e, rel)
		return
	}
	e[rel] = append(ResourceCollection(nil), r...)
}

// Get returns the embedded resources associated with the given relation.
func (e Embedded) Get(rel Relation) ResourceCollection {
	return append(ResourceCollection(nil), e[rel]...)
}

// Del deletes the resources associated with the given relation.
func (e Embedded) Del(rel Relation) {
	delete(e, rel)
}

// NewResource creates a Resource object for a given struct
// and its link.
func NewResource(p interface{}, selfUri string) *Resource {
	var r Resource

	r.Payload = p
	r.Flavor = DefaultFlavor

	if selfUri != "" {
		r.Links = make(LinkRelations)
		r.AddNewLink("self", selfUri)
	}

	// Embedded and Curies are lazily initialized on first use to avoid
	// allocating maps that most resources never populate.
	return &r
}

// AddLinkCollection appends a LinkCollection to the resource.
// l should be a LinkCollection
func (r *Resource) AddLinkCollection(rel Relation, l LinkCollection) {
	if r.Links == nil {
		r.Links = make(LinkRelations)
	}

	n := r.Links[rel]
	if n == nil {
		// new link
		r.Links[rel] = l
		return
	}

	if nc, ok := n.(LinkCollection); ok {
		r.Links[rel] = append(nc, l...)
		return
	}

	if nl, ok := n.(Link); ok {
		// prepend existing link to collection
		r.Links[rel] = append(LinkCollection{nl}, l...)
	}
}

// AddLink appends a Link to the resource.
// l should be a Link
func (r *Resource) AddLink(rel Relation, l Link) {
	if r.Links == nil {
		r.Links = make(LinkRelations)
	}

	n := r.Links[rel]
	if n == nil {
		// new link
		r.Links[rel] = l
		return
	}

	if nc, ok := n.(LinkCollection); ok {
		r.Links[rel] = append(nc, l)
		return
	}

	if nl, ok := n.(Link); ok {
		r.Links[rel] = append(LinkCollection{nl}, l)
		return
	}

	// something went wrong.. replace what is there with what is new
	r.Links[rel] = LinkCollection{l}
}

// AddNewLink appends a new Link object based on
// the rel and href params.
func (r *Resource) AddNewLink(rel Relation, href string) {
	r.AddLink(rel, NewLink(href, nil))
}

// RegisterCurie adds a Link relation of type 'curies' and returns a CurieHandle
// to allow users to fluently add new links that have this curie relation definition
func (r *Resource) RegisterCurie(name, href string, templated bool) *CurieHandle {
	l := LinkCollection{
		NewLink(href, LinkAttr{"name": name}, LinkAttr{"templated": templated}),
	}
	r.AddLinkCollection("curies", l)

	handle := &CurieHandle{Name: name, Resource: r}

	if r.Curies == nil {
		r.Curies = make(map[string]*CurieHandle)
	}
	r.Curies[name] = handle
	return handle
}

// Embed appends a Resource to the array of
// embedded resources.
// re should be a pointer to a Resource
func (r *Resource) Embed(rel Relation, re *Resource) {
	if r.Embedded == nil {
		r.Embedded = make(Embedded)
	}
	r.Embedded.Add(rel, re)
}

// EmbedCollection appends a ResourceCollection to the array of
// embedded resources.
// re should be a ResourceCollection
func (r *Resource) EmbedCollection(rel Relation, re ResourceCollection) {
	if r.Embedded == nil {
		r.Embedded = make(Embedded)
	}
	r.Embedded.AddCollection(rel, re)
}

// GetMap implements the interface Mapper.
func (r Resource) GetMap() Entry {
	mp := r.payloadMap()
	linksKey, embeddedKey := r.Flavor.keys()

	if r.Flavor == FlavorViaplay {
		// FlavorViaplay nests payload fields under "data" instead of
		// flattening them onto the resource object (specs/viaplay-hateoas.md #4).
		mapped := make(Entry, 3)

		if len(mp) > 0 {
			mapped["data"] = mp
		}

		if len(r.Links) > 0 {
			mapped[linksKey] = r.Links
		}

		if len(r.Embedded) > 0 {
			mapped[embeddedKey] = r.Embedded
		}

		return mapped
	}

	// Flattened flavors (FlavorDefault, FlavorHAL): add links/embedded
	// directly into mp instead of copying into a new map. This assumes mp is
	// always freshly allocated per call (true for getPayloadMap, and for any
	// well-behaved Mapper.GetMap implementation returning a literal/fresh
	// Entry rather than a shared/cached map instance across calls).
	if mp == nil {
		mp = make(Entry, 2)
	}

	if len(r.Links) > 0 {
		mp[linksKey] = r.Links
	}

	if len(r.Embedded) > 0 {
		mp[embeddedKey] = r.Embedded
	}

	return mp
}

// payloadMap resolves the resource's payload into an Entry, preferring the
// Mapper interface when the payload implements it and falling back to
// reflection-based field mapping otherwise.
func (r Resource) payloadMap() Entry {
	if mapper, ok := r.Payload.(Mapper); ok {
		return mapper.GetMap()
	}
	return r.getPayloadMap()
}

// payloadField is a cached description of a struct field that should be
// serialized as part of the payload map.
type payloadField struct {
	name  string
	index int
}

// fieldCache memoizes the reflected field layout of payload struct types so
// reflection (and json tag parsing) only happens once per concrete type.
var fieldCache sync.Map // map[reflect.Type][]payloadField

// cachedPayloadFields returns the payload field descriptors for t, computing
// and caching them on first use. The returned slice MUST NOT be mutated.
func cachedPayloadFields(t reflect.Type) []payloadField {
	if v, ok := fieldCache.Load(t); ok {
		if fields, ok := v.([]payloadField); ok {
			return fields
		}
	}

	n := t.NumField()
	fields := make([]payloadField, 0, n)
	for i := 0; i < n; i++ {
		typeField := t.Field(i)
		if !typeField.IsExported() {
			continue
		}

		tagValue := typeField.Tag.Get("json")
		// Strip the ",omitempty" modifier (semantics intentionally preserved:
		// the field is always emitted; only the modifier is removed from the key).
		if idx := strings.Index(tagValue, ",omitempty"); idx >= 0 {
			tagValue = tagValue[:idx] + tagValue[idx+len(",omitempty"):]
		}

		if tagValue == "-" {
			continue
		}

		if tagValue == "" {
			tagValue = typeField.Name
		}

		fields = append(fields, payloadField{name: tagValue, index: i})
	}

	actual, _ := fieldCache.LoadOrStore(t, fields)
	if cached, ok := actual.([]payloadField); ok {
		return cached
	}
	return fields
}

func (r *Resource) getPayloadMap() Entry {
	if r.Payload == nil {
		return Entry{}
	}

	val := reflect.ValueOf(r.Payload)
	// Dereference pointers so callers can pass either T or *T.
	for val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return Entry{}
		}
		val = val.Elem()
	}

	// Only structs have fields; anything else (map, slice, primitive) is ignored.
	if val.Kind() != reflect.Struct {
		return Entry{}
	}

	fields := cachedPayloadFields(val.Type())
	payloadMap := make(Entry, len(fields))
	for _, f := range fields {
		payloadMap[f.name] = val.Field(f.index).Interface()
	}

	return payloadMap
}

// MarshalJSON is a Marshaler implementation for Embedded.
// HAL permits a single embedded resource to be serialized as a single object
// and multiple resources as an array.
func (e Embedded) MarshalJSON() ([]byte, error) {
	if len(e) == 0 {
		return []byte("{}"), nil
	}

	// Single marshal pass: let the standard encoder pick the shape (object
	// vs array) directly from the Go value's type, instead of marshaling
	// each relation individually into an intermediate json.RawMessage and
	// then marshaling the resulting map a second time.
	items := make(map[Relation]any, len(e))
	for rel, resources := range e {
		switch len(resources) {
		case 0:
			continue
		case 1:
			items[rel] = resources[0]
		default:
			items[rel] = resources
		}
	}

	return json.Marshal(items)
}

// MarshalJSON is a Marshaler interface implementation
// for Resource struct
func (r Resource) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.GetMap())
}

// NewLink returns a new Link object.
func NewLink(href string, attrs ...LinkAttr) Link {
	l := make(Link)

	l["href"] = href

	for _, attr := range attrs {
		for k, v := range attr {
			l[k] = v
		}
	}

	return l
}
