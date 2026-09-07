package hal

import (
	"encoding/json"
	"testing"
)

type flavorDummy struct {
	Name string `json:"name"`
}

func TestFlavorDefaultViaplayOutOfTheBox(t *testing.T) {
	// NewResource inherits DefaultFlavor, which is FlavorViaplay: payload
	// nested under "data".
	expected := `{"data":{"name":"Dummy"},"links":{"self":{"href":"uri"}}}`

	r := NewResource(flavorDummy{"Dummy"}, "uri")

	jr, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}

func TestFlavorDefaultExplicitLegacyShape(t *testing.T) {
	// Explicitly opting into FlavorDefault (the zero value) restores the
	// pre-v2 flattened/unprefixed shape.
	expected := `{"links":{"self":{"href":"uri"}},"name":"Dummy"}`

	r := NewResource(flavorDummy{"Dummy"}, "uri")
	r.Flavor = FlavorDefault

	jr, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}

func TestFlavorHAL(t *testing.T) {
	expected := `{"_links":{"self":{"href":"uri"}},"name":"Dummy"}`

	r := NewResource(flavorDummy{"Dummy"}, "uri")
	r.Flavor = FlavorHAL

	jr, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}

func TestFlavorHALWithEmbedded(t *testing.T) {
	expected := `{"_embedded":{"child":{"_links":{"self":{"href":"uri2"}},"name":"Child"}},"_links":{"self":{"href":"uri"}},"name":"Dummy"}`

	r := NewResource(flavorDummy{"Dummy"}, "uri")
	r.Flavor = FlavorHAL

	child := NewResource(flavorDummy{"Child"}, "uri2")
	child.Flavor = FlavorHAL
	r.Embed("child", child)

	jr, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}

func TestFlavorViaplay(t *testing.T) {
	expected := `{"data":{"name":"Dummy"},"links":{"self":{"href":"uri"}}}`

	r := NewResource(flavorDummy{"Dummy"}, "uri")
	r.Flavor = FlavorViaplay

	jr, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}

func TestFlavorViaplayWithEmbeddedAndNoPayload(t *testing.T) {
	expected := `{"embedded":{"child":{"data":{"name":"Child"},"links":{"self":{"href":"uri2"}}}},"links":{"self":{"href":"uri"}}}`

	r := NewResource(struct{}{}, "uri")
	r.Flavor = FlavorViaplay

	child := NewResource(flavorDummy{"Child"}, "uri2")
	child.Flavor = FlavorViaplay
	r.Embed("child", child)

	jr, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}

func TestDefaultFlavorGlobal(t *testing.T) {
	old := DefaultFlavor
	defer func() { DefaultFlavor = old }()

	DefaultFlavor = FlavorHAL
	r := NewResource(flavorDummy{"Dummy"}, "uri")

	if r.Flavor != FlavorHAL {
		t.Fatalf("expected new resources to inherit DefaultFlavor, got %v", r.Flavor)
	}

	jr, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("%s", err)
	}
	expected := `{"_links":{"self":{"href":"uri"}},"name":"Dummy"}`
	if string(jr) != expected {
		t.Errorf("- Given:    %s\n- Expected: %s", jr, expected)
	}
}
