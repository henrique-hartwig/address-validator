package services

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/henrique/address-validator/internal/models"
)

type ValidatorService struct {
	geocodingService *GeocodingService
	cache            Cache
}

func NewValidatorService(geocodingService *GeocodingService, cache Cache) *ValidatorService {
	return &ValidatorService{
		geocodingService: geocodingService,
		cache:            cache,
	}
}

func (s *ValidatorService) ValidateAddress(ctx context.Context, address string) (*models.ValidateAddressResponse, error) {
	normalized := s.normalizeInput(address)

	cacheKey := s.generateCacheKey(normalized.Normalized)
	if cached, found := s.cache.Get(cacheKey); found {
		if result, ok := cached.(*models.ValidateAddressResponse); ok {
			return result, nil
		}
	}

	geocodingResult, err := s.geocodingService.Geocode(ctx, normalized.Normalized)
	if err != nil {
		return &models.ValidateAddressResponse{
			Status: "error",
			Error:  fmt.Sprintf("Failed to validate address: %v", err),
		}, nil
	}

	response := &models.ValidateAddressResponse{
		Status:      "success",
		Data:        geocodingResult.AddressData,
		Corrections: normalized.Changes,
	}

	s.cache.Set(cacheKey, response)

	return response, nil
}

func (s *ValidatorService) NormalizeInput(input string) *models.NormalizedInput {
	return s.normalizeInput(input)
}

func (s *ValidatorService) normalizeInput(input string) *models.NormalizedInput {
	original := input
	normalized := strings.TrimSpace(input)
	changes := []string{}

	words := strings.Fields(normalized)
	for i, word := range words {
		lower := strings.ToLower(word)

		if expansion, exists := StreetAbbreviations[lower]; exists {
			oldWord := word
			words[i] = expansion
			changes = append(changes, fmt.Sprintf("%s → %s", oldWord, expansion))
			continue
		}

		if expansion, exists := DirectionAbbreviations[lower]; exists {
			oldWord := word
			words[i] = expansion
			changes = append(changes, fmt.Sprintf("%s → %s", oldWord, expansion))
			continue
		}
	}
	normalized = strings.Join(words, " ")

	words = strings.Fields(normalized)
	for i, word := range words {
		lower := strings.ToLower(strings.TrimRight(word, ",."))

		if len(lower) < 4 || isNumeric(lower) {
			continue
		}

		if match, found := FindClosestMatch(lower, CommonStreetTypes, 2); found {
			if lower != match && !isCommonWord(lower) {
				suffix := ""
				if strings.HasSuffix(word, ",") {
					suffix = ","
				} else if strings.HasSuffix(word, ".") {
					suffix = "."
				}
				words[i] = match + suffix
				changes = append(changes, fmt.Sprintf("%s → %s (typo correction)", word, match+suffix))
				continue
			}
		}

		if len(lower) > 5 {
			if match, found := FindClosestMatch(lower, CommonCityNames, 2); found {
				if lower != match {
					suffix := ""
					if strings.HasSuffix(word, ",") {
						suffix = ","
					}
					words[i] = match + suffix
					changes = append(changes, fmt.Sprintf("%s → %s (city correction)", word, match+suffix))
				}
			}
		}
	}
	normalized = strings.Join(words, " ")

	words = strings.Fields(normalized)
	for i, word := range words {
		lower := strings.ToLower(strings.TrimRight(word, ",."))

		isAfterComma := i > 0 && strings.HasSuffix(words[i-1], ",")
		isTwoLetters := len(lower) == 2
		isAtEnd := i == len(words)-1

		isLikelyState := (isTwoLetters && (isAfterComma || isAtEnd)) ||
			(isAfterComma && len(lower) > 3)

		if isLikelyState {
			if stateAbbr, found := NormalizeUSState(lower); found {
				if !strings.EqualFold(word, stateAbbr) {
					words[i] = stateAbbr
					changes = append(changes, fmt.Sprintf("%s → %s (state)", word, stateAbbr))
				}
			}
		}
	}
	normalized = strings.Join(words, " ")

	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")
	normalized = strings.TrimSpace(normalized)

	return &models.NormalizedInput{
		Original:   original,
		Normalized: normalized,
		Changes:    changes,
	}
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"main": true, "park": true, "oak": true, "pine": true, "maple": true,
		"elm": true, "cedar": true, "lake": true, "hill": true, "view": true,
		"center": true, "first": true, "second": true, "third": true,
		"north": true, "south": true, "east": true, "west": true,
		"new": true, "old": true, "grand": true, "high": true, "spring": true,
	}
	return commonWords[strings.ToLower(word)]
}

func (s *ValidatorService) generateCacheKey(address string) string {
	hash := md5.Sum([]byte(strings.ToLower(address)))
	return "addr:" + hex.EncodeToString(hash[:])
}
