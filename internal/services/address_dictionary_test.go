package services

import (
	"testing"
)

func TestFindClosestMatch(t *testing.T) {
	tests := []struct {
		name        string
		word        string
		dictionary  []string
		maxDistance int
		wantMatch   string
		wantFound   bool
	}{
		{
			name:        "Typo in 'street'",
			word:        "stret",
			dictionary:  CommonStreetTypes,
			maxDistance: 2,
			wantMatch:   "street",
			wantFound:   true,
		},
		{
			name:        "Typo in 'avenue'",
			word:        "avenu",
			dictionary:  CommonStreetTypes,
			maxDistance: 2,
			wantMatch:   "avenue",
			wantFound:   true,
		},
		{
			name:        "Typo in 'boulevard'",
			word:        "boulevrd",
			dictionary:  CommonStreetTypes,
			maxDistance: 2,
			wantMatch:   "boulevard",
			wantFound:   true,
		},
		{
			name:        "Word without match",
			word:        "xyz123",
			dictionary:  CommonStreetTypes,
			maxDistance: 2,
			wantMatch:   "",
			wantFound:   false,
		},
		{
			name:        "City with typo - 'fransisco'",
			word:        "fransisco",
			dictionary:  CommonCityNames,
			maxDistance: 2,
			wantMatch:   "francisco",
			wantFound:   true,
		},
		{
			name:        "City with typo - 'angels' to 'angeles'",
			word:        "angels",
			dictionary:  CommonCityNames,
			maxDistance: 2,
			wantMatch:   "",
			wantFound:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatch, gotFound := FindClosestMatch(tt.word, tt.dictionary, tt.maxDistance)

			if gotFound != tt.wantFound {
				t.Errorf("FindClosestMatch() found = %v, want %v", gotFound, tt.wantFound)
			}

			if gotMatch != tt.wantMatch {
				t.Errorf("FindClosestMatch() match = %v, want %v", gotMatch, tt.wantMatch)
			}
		})
	}
}

func TestIsValidUSState(t *testing.T) {
	tests := []struct {
		name  string
		state string
		want  bool
	}{
		{"Abbreviation valid - CA", "CA", true},
		{"Abbreviation valid lowercase - ca", "ca", true},
		{"Full name - California", "california", true},
		{"Full name uppercase - California", "California", true},
		{"Invalid state", "XX", false},
		{"Invalid state", "Atlantis", false},
		{"New York", "new york", true},
		{"NY", "NY", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidUSState(tt.state); got != tt.want {
				t.Errorf("IsValidUSState(%v) = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}

func TestNormalizeUSState(t *testing.T) {
	tests := []struct {
		name      string
		state     string
		wantAbbr  string
		wantFound bool
	}{
		{
			name:      "Abbreviation CA",
			state:     "CA",
			wantAbbr:  "CA",
			wantFound: true,
		},
		{
			name:      "Full name California",
			state:     "california",
			wantAbbr:  "CA",
			wantFound: true,
		},
		{
			name:      "Typo - Californa",
			state:     "californa",
			wantAbbr:  "CA",
			wantFound: true,
		},
		{
			name:      "New York",
			state:     "new york",
			wantAbbr:  "NY",
			wantFound: true,
		},
		{
			name:      "Typo - Texs",
			state:     "texs",
			wantAbbr:  "TX",
			wantFound: true,
		},
		{
			name:      "Invalid state",
			state:     "InvalidState",
			wantAbbr:  "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAbbr, gotFound := NormalizeUSState(tt.state)

			if gotFound != tt.wantFound {
				t.Errorf("NormalizeUSState(%v) found = %v, want %v", tt.state, gotFound, tt.wantFound)
			}

			if gotAbbr != tt.wantAbbr {
				t.Errorf("NormalizeUSState(%v) abbr = %v, want %v", tt.state, gotAbbr, tt.wantAbbr)
			}
		})
	}
}
