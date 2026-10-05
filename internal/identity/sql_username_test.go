package identity

import "testing"

func TestDatabaseUsernamePreservesCompatibleName(t *testing.T) {
	if got := DatabaseUsername("Erick.Admin-1", "cluster."); got != "cluster.Erick.Admin-1" {
		t.Fatalf("got %q", got)
	}
}

func TestDatabaseUsernameMapsInvalidNameDeterministically(t *testing.T) {
	first := DatabaseUsername("usuario con espacios y un nombre demasiado largo", "3anrsu2PBeYSrxG.")
	second := DatabaseUsername("usuario con espacios y un nombre demasiado largo", "3anrsu2PBeYSrxG.")
	if first != second || len(first) != 32 || first[:20] != "3anrsu2PBeYSrxG.app_" {
		t.Fatalf("got %q and %q", first, second)
	}
}

func TestDatabaseUsernameHashesCompatibleNameThatDoesNotFitAfterPrefix(t *testing.T) {
	got := DatabaseUsername("sixteen_chars_okx", "3anrsu2PBeYSrxG.")
	if len(got) != 32 || got[:20] != "3anrsu2PBeYSrxG.app_" {
		t.Fatalf("got %q", got)
	}
}
