package catalog

import (
	"testing"
)

func TestDeduplicateByExternalID(t *testing.T) {
	id1 := "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d"
	id2 := "b2c3d4e5-f6a7-4b5c-9d0e-1f2a3b4c5d6e"
	id3 := "c3d4e5f6-a7b8-4c5d-0e1f-2a3b4c5d6e7f"

	tests := []struct {
		name    string
		input   []ProductEntry
		wantLen int
		wantIDs []string
	}{
		{
			name:    "empty input returns empty slice",
			input:   []ProductEntry{},
			wantLen: 0,
			wantIDs: []string{},
		},
		{
			name: "no duplicates — all entries kept",
			input: []ProductEntry{
				{Id: id1, Name: "A"},
				{Id: id2, Name: "B"},
				{Id: id3, Name: "C"},
			},
			wantLen: 3,
			wantIDs: []string{id1, id2, id3},
		},
		{
			name: "one duplicate — first occurrence kept",
			input: []ProductEntry{
				{Id: id1, Name: "First"},
				{Id: id1, Name: "Duplicate"},
				{Id: id2, Name: "Other"},
			},
			wantLen: 2,
			wantIDs: []string{id1, id2},
		},
		{
			name: "all entries share same ID — only first kept",
			input: []ProductEntry{
				{Id: id1, Name: "First"},
				{Id: id1, Name: "Second"},
				{Id: id1, Name: "Third"},
			},
			wantLen: 1,
			wantIDs: []string{id1},
		},
		{
			name: "first occurrence is preserved when duplicated",
			input: []ProductEntry{
				{Id: id1, Name: "original"},
				{Id: id1, Name: "copy"},
			},
			wantLen: 1,
			wantIDs: []string{id1},
			// extra check done below to confirm the Name of the kept entry
		},
		{
			name: "single entry is kept as-is",
			input: []ProductEntry{
				{Id: id1, Name: "Solo"},
			},
			wantLen: 1,
			wantIDs: []string{id1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := deduplicateByExternalID(tc.input)

			if len(got) != tc.wantLen {
				t.Errorf("len: got %d, want %d", len(got), tc.wantLen)
			}

			for i, wantID := range tc.wantIDs {
				if i >= len(got) {
					break
				}
				if got[i].Id != wantID {
					t.Errorf("entry[%d].Id: got %q, want %q", i, got[i].Id, wantID)
				}
			}

			// Verify first-occurrence semantics for the "first occurrence is preserved" case
			if tc.name == "first occurrence is preserved when duplicated" && len(got) == 1 {
				if got[0].Name != "original" {
					t.Errorf("expected first occurrence to be kept, got Name=%q", got[0].Name)
				}
			}
		})
	}
}
