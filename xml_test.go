package hal

import (
	"encoding/xml"
	"testing"
)

type xmlDummyStruct struct {
	Name string `json:"name"`
}

func TestResourceMarshalXML(t *testing.T) {
	expected := `<resource href="uri"><name>Dummy</name></resource>`

	r := NewResource(xmlDummyStruct{"Dummy"}, "uri")

	b, err := xml.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if string(b) != expected {
		t.Fatalf("- Given:    %s\n- Expected: %s", b, expected)
	}
}

func TestResourceMarshalXMLWithLinks(t *testing.T) {
	expected := `<resource href="uri"><link rel="help" href="/docs"></link><name>Dummy</name></resource>`

	r := NewResource(xmlDummyStruct{"Dummy"}, "uri")
	r.AddNewLink("help", "/docs")

	b, err := xml.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if string(b) != expected {
		t.Fatalf("- Given:    %s\n- Expected: %s", b, expected)
	}
}

func TestResourceMarshalXMLWithEmbedded(t *testing.T) {
	expected := `<resource href="uri"><name>Dummy</name>` +
		`<resource rel="child" href="uri2"><name>DummyEmbed</name></resource></resource>`

	r := NewResource(xmlDummyStruct{"Dummy"}, "uri")
	child := NewResource(xmlDummyStruct{"DummyEmbed"}, "uri2")
	r.Embed("child", child)

	b, err := xml.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if string(b) != expected {
		t.Fatalf("- Given:    %s\n- Expected: %s", b, expected)
	}
}

func TestResourceMarshalXMLWithMultipleEmbedded(t *testing.T) {
	expected := `<resource href="uri"><name>Dummy</name>` +
		`<resource rel="child" href="uri2"><name>DummyEmbed</name></resource>` +
		`<resource rel="child" href="uri3"><name>DummyEmbed2</name></resource></resource>`

	r := NewResource(xmlDummyStruct{"Dummy"}, "uri")
	r.Embed("child", NewResource(xmlDummyStruct{"DummyEmbed"}, "uri2"))
	r.Embed("child", NewResource(xmlDummyStruct{"DummyEmbed2"}, "uri3"))

	b, err := xml.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if string(b) != expected {
		t.Fatalf("- Given:    %s\n- Expected: %s", b, expected)
	}
}

type xmlNestedPayload struct {
	Tags []string
	Meta Entry
}

func (p xmlNestedPayload) GetMap() Entry {
	return Entry{
		"tags": p.Tags,
		"meta": p.Meta,
	}
}

func TestResourceMarshalXMLNestedMapAndSlice(t *testing.T) {
	expected := `<resource><meta><count>2</count></meta><tags>a</tags><tags>b</tags></resource>`

	r := NewResource(xmlNestedPayload{
		Tags: []string{"a", "b"},
		Meta: Entry{"count": 2},
	}, "")

	b, err := xml.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if string(b) != expected {
		t.Fatalf("- Given:    %s\n- Expected: %s", b, expected)
	}
}

func TestSanitizeXMLName(t *testing.T) {
	cases := map[string]string{
		"name":      "name",
		"full name": "full_name",
		"2fast":     "_2fast",
		"":          "_",
	}

	for in, want := range cases {
		if got := sanitizeXMLName(in); got != want {
			t.Errorf("sanitizeXMLName(%q) = %q, want %q", in, got, want)
		}
	}
}
