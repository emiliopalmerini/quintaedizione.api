package shared

import "testing"

func TestEscapeLike(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no special chars", "hello", "hello"},
		{"percent", "100%", `100\%`},
		{"underscore", "foo_bar", `foo\_bar`},
		{"backslash", `a\b`, `a\\b`},
		{"all special chars", `a%b_c\d`, `a\%b\_c\\d`},
		{"empty string", "", ""},
		{"only special chars", `%_\`, `\%\_\\`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeLike(tt.input)
			if got != tt.expected {
				t.Errorf("EscapeLike(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
