package maintenance

import "testing"

func TestValidTypes(t *testing.T) {
	for k := range AllTypes {
		if !IsValidType(k) {
			t.Fatalf("Unexpected invalid type detected: %s", k)
		}
	}

	if IsValidType("") {
		t.Fatalf("Unexpected valid empty type detected")
	}
	if IsValidType("invalidType") {
		t.Fatalf("Unexpected valid type detected: `invalidType`")
	}
}
