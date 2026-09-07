// Package hal XML support.
//
// This file implements a HAL+XML style Marshaler for Resource so that the
// same Resource graph used for JSON output can also be serialized as XML,
// without requiring callers to define separate XML-specific structs.
//
// The produced shape follows the conventions used by most HAL+XML
// implementations:
//
//	<resource href="/products/1">
//	  <link rel="help" href="/docs"/>
//	  <name>Some Product</name>
//	  <price>10</price>
//	  <resource rel="category" href="/categories/1">
//	    <name>Some Category</name>
//	  </resource>
//	</resource>
package hal

import (
	"encoding/xml"
	"fmt"
	"reflect"
	"regexp"
	"sort"
)

// hrefKey is the well-known Link/LinkAttr map key holding the link's URI.
const hrefKey = "href"

// invalidXMLNameChars matches characters that are not valid in an XML
// element local name once the leading-character rules have been applied.
var invalidXMLNameChars = regexp.MustCompile(`[^A-Za-z0-9_.-]`)

// sanitizeXMLName converts an arbitrary map/field key into a safe XML
// element local name: invalid characters are replaced with "_" and a
// leading digit is prefixed with "_" since XML names cannot start with one.
func sanitizeXMLName(name string) string {
	if name == "" {
		return "_"
	}

	name = invalidXMLNameChars.ReplaceAllString(name, "_")

	if name[0] >= '0' && name[0] <= '9' {
		name = "_" + name
	}

	return name
}

// selfHref returns the resource's self link href, if any.
func (r Resource) selfHref() string {
	v, ok := r.Links["self"]
	if !ok {
		return ""
	}

	switch val := v.(type) {
	case Link:
		if href, ok := val[hrefKey].(string); ok {
			return href
		}
	case LinkCollection:
		if len(val) > 0 {
			if href, ok := val[0][hrefKey].(string); ok {
				return href
			}
		}
	}

	return ""
}

// linksForRelation normalizes a LinkRelations value (which may be a single
// Link or a LinkCollection) into a slice of Link for uniform iteration.
func linksForRelation(v any) []Link {
	switch val := v.(type) {
	case Link:
		return []Link{val}
	case LinkCollection:
		return val
	default:
		return nil
	}
}

// encodeLinksXML writes a <link rel="..." href="..." .../> element for every
// link relation except "self", which is instead encoded as the href
// attribute of the enclosing <resource> element.
func (r Resource) encodeLinksXML(enc *xml.Encoder) error {
	if len(r.Links) == 0 {
		return nil
	}

	rels := make([]Relation, 0, len(r.Links))
	for rel := range r.Links {
		if rel == "self" {
			continue
		}
		rels = append(rels, rel)
	}
	sort.Slice(rels, func(i, j int) bool { return rels[i] < rels[j] })

	for _, rel := range rels {
		for _, l := range linksForRelation(r.Links[rel]) {
			if err := encodeLinkXML(enc, rel, l); err != nil {
				return err
			}
		}
	}

	return nil
}

func encodeLinkXML(enc *xml.Encoder, rel Relation, l Link) error {
	start := xml.StartElement{Name: xml.Name{Local: "link"}}
	start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "rel"}, Value: string(rel)})

	if href, ok := l[hrefKey].(string); ok {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: hrefKey}, Value: href})
	}

	keys := make([]string, 0, len(l))
	for k := range l {
		if k == hrefKey {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: sanitizeXMLName(k)}, Value: fmt.Sprint(l[k])})
	}

	if err := enc.EncodeToken(start); err != nil {
		return err
	}
	return enc.EncodeToken(start.End())
}

// encodePayloadXML writes one child element per payload entry, recursing
// into nested maps and slices since encoding/xml cannot natively marshal
// map[string]any values.
func (r Resource) encodePayloadXML(enc *xml.Encoder) error {
	mp := r.payloadMap()
	keys := make([]string, 0, len(mp))
	for k := range mp {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if err := encodeXMLValue(enc, k, mp[k]); err != nil {
			return err
		}
	}

	return nil
}

// encodeEmbeddedXML writes a <resource rel="..." href="..."> element for
// every embedded resource, preserving the collection order.
func (r Resource) encodeEmbeddedXML(enc *xml.Encoder) error {
	if len(r.Embedded) == 0 {
		return nil
	}

	rels := make([]Relation, 0, len(r.Embedded))
	for rel := range r.Embedded {
		rels = append(rels, rel)
	}
	sort.Slice(rels, func(i, j int) bool { return rels[i] < rels[j] })

	for _, rel := range rels {
		for _, child := range r.Embedded[rel] {
			if child == nil {
				continue
			}

			start := xml.StartElement{
				Name: xml.Name{Local: "resource"},
				Attr: []xml.Attr{{Name: xml.Name{Local: "rel"}, Value: string(rel)}},
			}
			if err := enc.EncodeElement(child, start); err != nil {
				return err
			}
		}
	}

	return nil
}

// MarshalXML implements xml.Marshaler for Resource, producing a HAL+XML
// representation: a <resource> element carrying the self href, <link>
// children for related links, one element per payload field, and nested
// <resource rel="..."> children for embedded resources.
func (r Resource) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: "resource"}

	if self := r.selfHref(); self != "" {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: hrefKey}, Value: self})
	}

	if err := enc.EncodeToken(start); err != nil {
		return err
	}

	if err := r.encodeLinksXML(enc); err != nil {
		return err
	}

	if err := r.encodePayloadXML(enc); err != nil {
		return err
	}

	if err := r.encodeEmbeddedXML(enc); err != nil {
		return err
	}

	return enc.EncodeToken(start.End())
}

// encodeXMLValue writes v as a child element named name, recursing into
// maps and slices as needed since they are not natively supported by
// encoding/xml when their element type is any/interface{}.
func encodeXMLValue(enc *xml.Encoder, name string, v any) error {
	switch val := v.(type) {
	case nil:
		return encodeEmptyXMLElement(enc, name)
	case Entry:
		return encodeXMLMap(enc, name, val)
	case map[string]any:
		return encodeXMLMap(enc, name, val)
	case xml.Marshaler:
		start := xml.StartElement{Name: xml.Name{Local: sanitizeXMLName(name)}}
		return enc.EncodeElement(val, start)
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Pointer:
		if rv.IsNil() {
			return encodeEmptyXMLElement(enc, name)
		}
		return encodeXMLValue(enc, name, rv.Elem().Interface())
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			if err := encodeXMLValue(enc, name, rv.Index(i).Interface()); err != nil {
				return err
			}
		}
		return nil
	case reflect.Map:
		m := make(map[string]any, rv.Len())
		for _, k := range rv.MapKeys() {
			m[fmt.Sprint(k.Interface())] = rv.MapIndex(k).Interface()
		}
		return encodeXMLMap(enc, name, m)
	}

	start := xml.StartElement{Name: xml.Name{Local: sanitizeXMLName(name)}}
	return enc.EncodeElement(fmt.Sprint(v), start)
}

func encodeXMLMap(enc *xml.Encoder, name string, m map[string]any) error {
	start := xml.StartElement{Name: xml.Name{Local: sanitizeXMLName(name)}}
	if err := enc.EncodeToken(start); err != nil {
		return err
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if err := encodeXMLValue(enc, k, m[k]); err != nil {
			return err
		}
	}

	return enc.EncodeToken(start.End())
}

func encodeEmptyXMLElement(enc *xml.Encoder, name string) error {
	start := xml.StartElement{Name: xml.Name{Local: sanitizeXMLName(name)}}
	if err := enc.EncodeToken(start); err != nil {
		return err
	}
	return enc.EncodeToken(start.End())
}
