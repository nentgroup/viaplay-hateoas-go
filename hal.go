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
	Entry map[string]interface{}

	// Mapper is the interface implemented by the objects
	// that can be converted into HAL format.
	Mapper interface {
		GetMap() Entry
	}

	Relation string

	CurieHandle struct {
		Name string
		*Resource
	}

	// LinkAttr types that store hyperlinks and its attributes.
	LinkAttr       map[string]interface{}
	Link           LinkAttr
	LinkCollection []Link
	LinkRelations  map[Relation]interface{}

	// Resource is a struct that stores a resource data.
	// It represents a converted object in the HAL spec by
	// containing all its fields and also a set of related links
	// and a sub-set of recursively related resources.
	Resource struct {
		Payload  interface{}
		Links    LinkRelations
		Embedded Embedded
		Curies   map[string]*CurieHandle
	}
	ResourceCollection []*Resource

	Embedded map[Relation]interface{}
)

// AddNewLink adds a link to the resources_link collection
// prepended with the curie Name
func (c CurieHandle) AddNewLink(rel Relation, href string) {
	rel = Relation(c.Name) + ":" + rel
	c.AddLink(rel, NewLink(href, nil))
}

// AddCollection appends the resource into the list of embedded
// resources with the specified relation.
// r should be  a ResourceCollection
func (e Embedded) AddCollection(rel Relation, r ResourceCollection) {
	n := e[rel]
	if n == nil {
		// new embed
		e[rel] = r
		return
	}

	if nc, ok := n.([]*Resource); ok {
		e[rel] = append(nc, r...)
		return
	}

	if nr, ok := n.(*Resource); ok {
		e[rel] = append([]*Resource{nr}, r...)
		return
	}
}

// Add appends the resource into the list of embedded
// resources with the specified relation.
// r should be a Resource
func (e Embedded) Add(rel Relation, r *Resource) {
	n := e[rel]
	if n == nil {
		// new embed
		e[rel] = []*Resource{r}
		return
	}

	if nec, ok := n.([]*Resource); ok {
		e[rel] = append(nec, r)
		return
	}

	if nee, ok := n.(*Resource); ok {
		e[rel] = append([]*Resource{nee}, r)
		return
	}

	// something went wrong.. replace what is there with what is new
	e[rel] = []*Resource{r}
}

// Set sets the resource into the list of embedded
// resources with the specified relation. It replaces
// any existing resources associated with the relation.
// r should be a pointer to a Resource
func (e Embedded) Set(rel Relation, r *Resource) {
	e[rel] = r
}

// SetCollection sets the resource into the list of embedded
// resources with the specified relation. It replaces
// any existing resources associated with the relation.
// r should be a ResourceCollection
func (e Embedded) SetCollection(rel Relation, r ResourceCollection) {
	e[rel] = r
}

// Get gets the resources associated with the
// given relation.
//func (e Embedded) Get(rel Relation) []*Resource {
//	return e[rel]
//}

// Del deletes the resources associated with the
// given relation.
func (e Embedded) Del(rel Relation) {
	delete(e, rel)
}

// NewResource creates a Resource object for a given struct
// and its link.
func NewResource(p interface{}, selfUri string) *Resource {
	var r Resource

	r.Payload = p

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
	var mp Entry
	// Check if payload implements Mapper interface
	if mapper, ok := r.Payload.(Mapper); ok {
		mp = mapper.GetMap()
	} else {
		mp = r.getPayloadMap()
	}

	// Pre-size to avoid map regrowth: payload entries + at most "links" + "embedded".
	mapped := make(Entry, len(mp)+2)

	for k, v := range mp {
		mapped[k] = v
	}

	if len(r.Links) > 0 {
		mapped["links"] = r.Links
	}

	if len(r.Embedded) > 0 {
		mapped["embedded"] = r.Embedded
	}

	return mapped
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
	for val.Kind() == reflect.Ptr {
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
