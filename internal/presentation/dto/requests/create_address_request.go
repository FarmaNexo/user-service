// internal/presentation/dto/requests/create_address_request.go
package requests

// CreateAddressRequest DTO para crear una dirección
type CreateAddressRequest struct {
	Label      string   `json:"label,omitempty"`
	Street     string   `json:"street"`
	City       string   `json:"city"`
	State      string   `json:"state,omitempty"`
	PostalCode string   `json:"postal_code,omitempty"`
	Country    string   `json:"country,omitempty"`
	IsDefault  bool     `json:"is_default"`
	Latitude   *float64 `json:"latitude,omitempty"`
	Longitude  *float64 `json:"longitude,omitempty"`
}
