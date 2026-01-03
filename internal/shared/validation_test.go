package shared

import (
	"testing"
)

func TestValidateID(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   string
		wantErr bool
	}{
		{"valid id", "id", "barbaro", false},
		{"valid with hyphen", "id", "sotto-classe", false},
		{"valid with underscore", "id", "some_id", false},
		{"valid with numbers", "id", "class123", false},
		{"empty value", "id", "", true},
		{"too long", "id", "aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeefffff", true},
		{"invalid chars space", "id", "invalid id", true},
		{"invalid chars special", "id", "invalid@id", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateID(tt.field, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
