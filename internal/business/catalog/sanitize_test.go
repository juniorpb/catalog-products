package catalog

import (
	"testing"

	"github.com/google/uuid"
)

func TestSanitizeEntries(t *testing.T) {
	validUUID := "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d"

	tests := []struct {
		name        string
		input       []ProductEntry
		wantLen     int
		checkOutput func(t *testing.T, got []ProductEntry)
	}{
		{
			name: "valid entry is kept unchanged",
			input: []ProductEntry{
				{Id: validUUID, SellerName: "Acme", Name: "Widget", Category: "Tools"},
			},
			wantLen: 1,
			checkOutput: func(t *testing.T, got []ProductEntry) {
				if got[0].Id != validUUID {
					t.Errorf("expected ID to be preserved, got %s", got[0].Id)
				}
			},
		},
		{
			name: "invalid UUID is replaced with a new valid UUID",
			input: []ProductEntry{
				{Id: "not-a-uuid", SellerName: "Acme", Name: "Widget", Category: "Tools"},
			},
			wantLen: 1,
			checkOutput: func(t *testing.T, got []ProductEntry) {
				if got[0].Id == "not-a-uuid" {
					t.Error("expected invalid UUID to be replaced")
				}
				if _, err := uuid.Parse(got[0].Id); err != nil {
					t.Errorf("replacement UUID is not valid: %s", got[0].Id)
				}
			},
		},
		{
			name: "entry with empty Name is discarded",
			input: []ProductEntry{
				{Id: validUUID, SellerName: "Acme", Name: "", Category: "Tools"},
			},
			wantLen: 0,
		},
		{
			name: "entry with whitespace-only Name is discarded after trim",
			input: []ProductEntry{
				{Id: validUUID, SellerName: "Acme", Name: "   ", Category: "Tools"},
			},
			wantLen: 0,
		},
		{
			name: "entry with empty SellerName is discarded",
			input: []ProductEntry{
				{Id: validUUID, SellerName: "", Name: "Widget", Category: "Tools"},
			},
			wantLen: 0,
		},
		{
			name: "leading and trailing spaces are trimmed from Name, SellerName and Category",
			input: []ProductEntry{
				{Id: validUUID, SellerName: "  Acme  ", Name: "  Widget  ", Category: "  Tools  "},
			},
			wantLen: 1,
			checkOutput: func(t *testing.T, got []ProductEntry) {
				if got[0].Name != "Widget" {
					t.Errorf("Name: got %q, want %q", got[0].Name, "Widget")
				}
				if got[0].SellerName != "Acme" {
					t.Errorf("SellerName: got %q, want %q", got[0].SellerName, "Acme")
				}
				if got[0].Category != "Tools" {
					t.Errorf("Category: got %q, want %q", got[0].Category, "Tools")
				}
			},
		},
		{
			name: "multiple entries — valid kept, invalid discarded or fixed",
			input: []ProductEntry{
				{Id: validUUID, SellerName: "Acme", Name: "Widget", Category: "Tools"},
				{Id: "bad-uuid", SellerName: "Acme", Name: "Gadget", Category: "Tech"},
				{Id: validUUID, SellerName: "", Name: "NoSeller", Category: "Tools"},
			},
			wantLen: 2,
		},
		{
			name:    "empty input returns empty slice",
			input:   []ProductEntry{},
			wantLen: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeEntries(tc.input)

			if len(got) != tc.wantLen {
				t.Errorf("len: got %d, want %d", len(got), tc.wantLen)
			}

			if tc.checkOutput != nil && len(got) > 0 {
				tc.checkOutput(t, got)
			}
		})
	}
}
