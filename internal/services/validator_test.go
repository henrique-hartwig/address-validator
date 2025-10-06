package services

import (
	"testing"
)

func TestNormalizeInput(t *testing.T) {
	cache := NewMockCacheService()
	geocodingService := NewGeocodingService(
		"test_key_a", "https://test.api",
		"test_key_b", "https://test.api",
		cache,
	)
	validatorService := NewValidatorService(geocodingService, cache)

	tests := []struct {
		name              string
		input             string
		shouldHaveChanges bool
	}{
		{
			name:              "Fix street typo",
			input:             "123 Main Stret",
			shouldHaveChanges: true,
		},
		{
			name:              "Expand abbreviations",
			input:             "456 Oak Ave",
			shouldHaveChanges: true,
		},
		{
			name:              "Multiple fixes",
			input:             "789 Park Blvd, San Fransisco",
			shouldHaveChanges: true,
		},
		{
			name:              "Already normalized",
			input:             "123 Main Street",
			shouldHaveChanges: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validatorService.normalizeInput(tt.input)

			if tt.shouldHaveChanges && len(result.Changes) == 0 {
				t.Errorf("Expected changes for input %v, but got none", tt.input)
			}

			if result.Normalized == "" {
				t.Error("Normalized output should not be empty")
			}

			if result.Original != tt.input {
				t.Errorf("Original should be preserved: got %v, want %v", result.Original, tt.input)
			}
		})
	}
}

func TestCacheKeyGeneration(t *testing.T) {
	cache := NewMockCacheService()
	geocodingService := NewGeocodingService("test_key_a", "https://test.api", "test_key_b", "https://test.api", cache)
	validatorService := NewValidatorService(geocodingService, cache)

	addr1 := "123 Main Street"
	addr2 := "123 Main Street"
	addr3 := "456 Oak Avenue"

	key1 := validatorService.generateCacheKey(addr1)
	key2 := validatorService.generateCacheKey(addr2)
	key3 := validatorService.generateCacheKey(addr3)

	if key1 != key2 {
		t.Errorf("Same address generated different cache keys: %v != %v", key1, key2)
	}

	if key1 == key3 {
		t.Errorf("Different addresses generated same cache key")
	}
}
