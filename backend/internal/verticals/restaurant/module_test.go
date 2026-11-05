package restaurant

import (
	"encoding/json"
	"testing"
)

func TestModuleManifestMatchesDefinition(t *testing.T) {
	var got struct {
		ID           string   `json:"id"`
		Capabilities []string `json:"capabilities"`
	}
	if err := json.Unmarshal(manifest, &got); err != nil {
		t.Fatal(err)
	}
	m := Module{}
	if got.ID != m.ID() {
		t.Fatalf("manifest id=%q, module id=%q", got.ID, m.ID())
	}
	if len(got.Capabilities) != len(m.Capabilities()) {
		t.Fatalf("manifest capabilities=%v, module capabilities=%v", got.Capabilities, m.Capabilities())
	}
	for i, capability := range m.Capabilities() {
		if got.Capabilities[i] != capability {
			t.Fatalf("manifest capability[%d]=%q, module capability=%q", i, got.Capabilities[i], capability)
		}
	}
}
