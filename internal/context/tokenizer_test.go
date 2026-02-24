package context

import "testing"

func TestTokenizer_Count(t *testing.T) {
	tok, err := NewTokenizer()
	if err != nil {
		t.Fatalf("NewTokenizer: %v", err)
	}

	count := tok.Count("Hello, world!")
	if count <= 0 {
		t.Errorf("expected positive token count, got %d", count)
	}
}

func TestTokenizer_Count_EmptyString(t *testing.T) {
	tok, err := NewTokenizer()
	if err != nil {
		t.Fatalf("NewTokenizer: %v", err)
	}

	count := tok.Count("")
	if count != 0 {
		t.Errorf("expected 0 tokens for empty string, got %d", count)
	}
}

func TestTokenizer_Truncate(t *testing.T) {
	tok, err := NewTokenizer()
	if err != nil {
		t.Fatalf("NewTokenizer: %v", err)
	}

	long := "This is a fairly long string that should have more than five tokens in total."
	truncated := tok.Truncate(long, 5)

	// Truncated should be shorter.
	if len(truncated) >= len(long) {
		t.Error("truncated string should be shorter than original")
	}

	// Verify the truncated string has <= 5 tokens.
	tokenCount := tok.Count(truncated)
	if tokenCount > 5 {
		t.Errorf("truncated to 5 tokens but Count says %d", tokenCount)
	}
}

func TestTokenizer_Truncate_ShortString(t *testing.T) {
	tok, err := NewTokenizer()
	if err != nil {
		t.Fatalf("NewTokenizer: %v", err)
	}

	short := "Hi"
	result := tok.Truncate(short, 100)
	if result != short {
		t.Errorf("short string should not be truncated: got %q", result)
	}
}
