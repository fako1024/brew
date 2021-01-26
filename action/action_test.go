package action

import "testing"

func TestValidTypes(t *testing.T) {
	for k, v := range categories {
		if category, isValid := Categorize(k); !isValid || category != v {
			t.Fatalf("Unexpected invalid type detected: %s", k)
		}
	}

	if category, isValid := Categorize(""); isValid || category != "" {
		t.Fatalf("Unexpected valid empty type detected")
	}
	if category, isValid := Categorize("invalidType"); isValid || category != "" {
		t.Fatalf("Unexpected valid empty type detected")
	}
}

func TestCategories(t *testing.T) {
	for k, v := range Categories() {
		if category, isValid := Categorize(k); !isValid || category != v {
			t.Fatalf("Unexpected invalid type detected: %s", k)
		}
	}
}
