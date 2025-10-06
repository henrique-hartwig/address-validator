package models

type ValidateAddressRequest struct {
	Address string `json:"address" binding:"required" example:"123 Main Stret, San Fransisco, CA, 94102"`
}

type ValidateAddressResponse struct {
	Status      string       `json:"status" example:"success"`
	Data        *AddressData `json:"data,omitempty"`
	Corrections []string     `json:"corrections,omitempty" example:"Stret → street (typo correction),Fransisco → francisco (city correction)"`
	Error       string       `json:"error,omitempty" example:"Failed to validate address"`
}

type AddressData struct {
	Street     string `json:"street" example:"Main Street"`
	Number     string `json:"number" example:"123"`
	City       string `json:"city" example:"San Francisco"`
	State      string `json:"state" example:"CA"`
	PostalCode string `json:"postal_code" example:"94102"`
	County     string `json:"county,omitempty" example:"San Francisco County"`
	Country    string `json:"country" example:"United States"`
	Formatted  string `json:"formatted" example:"123 Main Street, San Francisco, CA 94102"`
}

type GeocodingResponse struct {
	Success     bool
	AddressData *AddressData
	Provider    string
	Error       error
}

type NormalizedInput struct {
	Original   string
	Normalized string
	Changes    []string
}
