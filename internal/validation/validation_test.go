package validation

import "testing"

func TestTextRejectsWhitespaceAndLongValues(t *testing.T) {
	if _, err := Text("   ", "nombre", 10, true); err == nil {
		t.Fatal("expected whitespace-only value to fail")
	}
	if _, err := Text("abcdefghijkl", "nombre", 10, true); err == nil {
		t.Fatal("expected long value to fail")
	}
	if got, err := Text("  creatina  ", "nombre", 20, true); err != nil || got != "creatina" {
		t.Fatalf("got %q, %v", got, err)
	}
}
func TestImageURLProtocols(t *testing.T) {
	for _, value := range []string{"javascript:alert(1)", "file:///secret", "https://"} {
		if ImageURL(value) == nil {
			t.Fatalf("expected %q to fail", value)
		}
	}
	if err := ImageURL("https://example.com/product.webp"); err != nil {
		t.Fatal(err)
	}
}
func TestClosedEnums(t *testing.T) {
	if Role("SUPERADMIN") == nil {
		t.Fatal("arbitrary role accepted")
	}
	if OrderStatus("PAID") == nil {
		t.Fatal("arbitrary status accepted")
	}
}
