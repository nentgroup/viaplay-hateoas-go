package hal

import (
	"encoding/json"
	"testing"
)

type genericsProduct struct {
	ID   int
	Name string
}

func TestPayload(t *testing.T) {
	r := NewResource(genericsProduct{ID: 1, Name: "Widget"}, "/products/1")

	p, ok := Payload[genericsProduct](r)
	if !ok {
		t.Fatalf("expected payload assertion to succeed")
	}
	if p.Name != "Widget" {
		t.Errorf("got %q, want %q", p.Name, "Widget")
	}

	if _, ok := Payload[string](r); ok {
		t.Errorf("expected payload assertion to fail for mismatched type")
	}
}

func TestWithMapper(t *testing.T) {
	expected := `{"data":{"name":"Widget"},"links":{"self":{"href":"/products/1"}}}`

	mapped := WithMapper(genericsProduct{ID: 1, Name: "Widget"}, func(p genericsProduct) Entry {
		return Entry{"name": p.Name}
	})

	r := NewResource(mapped, "/products/1")

	jr, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}

func TestNewResourceCollection(t *testing.T) {
	products := []genericsProduct{
		{ID: 1, Name: "Widget"},
		{ID: 2, Name: "Gadget"},
	}

	rc := NewResourceCollection(products, func(p genericsProduct) string {
		return "/products/" + string(rune('0'+p.ID))
	})

	if len(rc) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(rc))
	}

	if got, _ := Payload[genericsProduct](rc[0]); got.Name != "Widget" {
		t.Errorf("got %q, want %q", got.Name, "Widget")
	}
	if got, _ := Payload[genericsProduct](rc[1]); got.Name != "Gadget" {
		t.Errorf("got %q, want %q", got.Name, "Gadget")
	}

	if rc[0].selfHref() != "/products/1" {
		t.Errorf("got %q, want %q", rc[0].selfHref(), "/products/1")
	}
}
