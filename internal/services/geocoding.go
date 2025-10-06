package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/henrique/address-validator/internal/models"
)

type GeocodingService struct {
	apiKeyA  string
	baseURLa string
	apiKeyB  string
	baseURLb string
	cache    Cache
	client   *http.Client
}

func NewGeocodingService(apiKeyA, baseURLa, apiKeyB, baseURLb string, cache Cache) *GeocodingService {
	return &GeocodingService{
		apiKeyA:  apiKeyA,
		baseURLa: baseURLa,
		apiKeyB:  apiKeyB,
		baseURLb: baseURLb,
		cache:    cache,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type GeoapifyResponse struct {
	Type     string            `json:"type"`
	Features []GeoapifyFeature `json:"features"`
	Query    GeoapifyQuery     `json:"query"`
}

type GeoapifyFeature struct {
	Type       string             `json:"type"`
	Properties GeoapifyProperties `json:"properties"`
	Geometry   GeoapifyGeometry   `json:"geometry"`
	Bbox       []float64          `json:"bbox"`
}

type GeoapifyProperties struct {
	Country      string  `json:"country"`
	CountryCode  string  `json:"country_code"`
	State        string  `json:"state"`
	StateCode    string  `json:"state_code"`
	County       string  `json:"county"`
	City         string  `json:"city"`
	Postcode     string  `json:"postcode"`
	Suburb       string  `json:"suburb"`
	Street       string  `json:"street"`
	HouseNumber  string  `json:"housenumber"`
	Formatted    string  `json:"formatted"`
	AddressLine1 string  `json:"address_line1"`
	AddressLine2 string  `json:"address_line2"`
	ResultType   string  `json:"result_type"`
	Lon          float64 `json:"lon"`
	Lat          float64 `json:"lat"`
}

type GeoapifyGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type GeoapifyQuery struct {
	Text   string                `json:"text"`
	Parsed GeoapifyParsedAddress `json:"parsed"`
}

type GeoapifyParsedAddress struct {
	HouseNumber  string `json:"housenumber"`
	Street       string `json:"street"`
	Postcode     string `json:"postcode"`
	District     string `json:"district"`
	City         string `json:"city"`
	Country      string `json:"country"`
	ExpectedType string `json:"expected_type"`
}

type SmartyResponse struct {
	Suggestions []SmartySuggestion `json:"suggestions"`
}

type SmartySuggestion struct {
	StreetLine string `json:"street_line"`
	Secondary  string `json:"secondary"`
	City       string `json:"city"`
	State      string `json:"state"`
	Zipcode    string `json:"zipcode"`
	Entries    int    `json:"entries"`
}

func (g *GeocodingService) Geocode(ctx context.Context, address string) (*models.GeocodingResponse, error) {
	if g.apiKeyA != "" && g.baseURLa != "" {
		result, err := g.geocodeWithGeoapify(ctx, address)
		if err == nil && result != nil && result.Success {
			return result, nil
		}
		if err != nil {
			fmt.Printf("Provider A (Geoapify) error: %v, trying fallback...\n", err)
		} else if result != nil && !result.Success {
			fmt.Printf("Provider A (Geoapify) returned no results, trying fallback...\n")
		}
	}

	if g.apiKeyB != "" && g.baseURLb != "" {
		result, err := g.geocodeWithSmarty(ctx, address)
		if err == nil && result != nil && result.Success {
			return result, nil
		}
		if err != nil {
			fmt.Printf("Provider B (Smarty) error: %v\n", err)
		} else if result != nil && !result.Success {
			fmt.Printf("Provider B (Smarty) returned no results\n")
		}
	}

	return &models.GeocodingResponse{
		Success:  false,
		Provider: "none",
		Error:    fmt.Errorf("all geocoding providers failed"),
	}, fmt.Errorf("failed to geocode address")
}

func (g *GeocodingService) geocodeWithGeoapify(ctx context.Context, address string) (*models.GeocodingResponse, error) {
	params := url.Values{}
	params.Add("text", address)
	params.Add("apiKey", g.apiKeyA)

	requestURL := fmt.Sprintf("%s?%s", g.baseURLa, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var geoapifyResp GeoapifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoapifyResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(geoapifyResp.Features) == 0 {
		return &models.GeocodingResponse{
			Success:  false,
			Provider: "geoapify",
			Error:    fmt.Errorf("no results found"),
		}, nil
	}

	feature := geoapifyResp.Features[0]
	props := feature.Properties

	addressData := &models.AddressData{
		Street:     formatStreetAddress(props.HouseNumber, props.Street),
		Number:     props.HouseNumber,
		City:       props.City,
		State:      props.StateCode,
		PostalCode: props.Postcode,
		County:     props.County,
		Country:    props.Country,
		Formatted:  props.Formatted,
	}

	return &models.GeocodingResponse{
		Success:     true,
		AddressData: addressData,
		Provider:    "geoapify",
		Error:       nil,
	}, nil
}

func (g *GeocodingService) geocodeWithSmarty(ctx context.Context, address string) (*models.GeocodingResponse, error) {
	params := url.Values{}
	params.Add("key", g.apiKeyB)
	params.Add("search", address)
	params.Add("max_results", "1")
	params.Add("license", "us-autocomplete-pro-cloud")

	requestURL := fmt.Sprintf("%s?%s", g.baseURLb, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Referer", "localhost:3000")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var smartyResp SmartyResponse
	if err := json.NewDecoder(resp.Body).Decode(&smartyResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(smartyResp.Suggestions) == 0 {
		return &models.GeocodingResponse{
			Success:  false,
			Provider: "smarty",
			Error:    fmt.Errorf("no results found"),
		}, nil
	}

	suggestion := smartyResp.Suggestions[0]

	number, street := parseStreetLine(suggestion.StreetLine)

	addressData := &models.AddressData{
		Street:     street,
		Number:     number,
		City:       suggestion.City,
		State:      suggestion.State,
		PostalCode: suggestion.Zipcode,
		County:     "",
		Country:    "United States",
		Formatted:  formatAddress(suggestion),
	}

	return &models.GeocodingResponse{
		Success:     true,
		AddressData: addressData,
		Provider:    "smarty",
		Error:       nil,
	}, nil
}

func parseStreetLine(streetLine string) (number string, street string) {
	parts := strings.Fields(streetLine)
	if len(parts) == 0 {
		return "", ""
	}

	firstPart := parts[0]
	if isHouseNumber(firstPart) {
		number = firstPart
		if len(parts) > 1 {
			street = strings.Join(parts[1:], " ")
		}
	} else {
		street = streetLine
	}

	return number, street
}

func isHouseNumber(s string) bool {
	for _, c := range s {
		if (c < '0' || c > '9') && c != '-' {
			return false
		}
	}
	return len(s) > 0
}

func formatAddress(s SmartySuggestion) string {
	parts := []string{s.StreetLine}

	if s.Secondary != "" {
		parts = append(parts, s.Secondary)
	}

	cityStateZip := fmt.Sprintf("%s, %s %s", s.City, s.State, s.Zipcode)
	parts = append(parts, cityStateZip)

	return strings.Join(parts, ", ")
}

func formatStreetAddress(number, street string) string {
	if number != "" && street != "" {
		return street
	}
	if street != "" {
		return street
	}
	return ""
}
