// internal/application/commands/create_address_command.go
package commands

// CreateAddressCommand representa el comando para crear una dirección
type CreateAddressCommand struct {
	UserID     string
	Label      string
	Street     string
	City       string
	State      string
	PostalCode string
	Country    string
	IsDefault  bool
	Latitude   *float64
	Longitude  *float64
}

func (c CreateAddressCommand) GetName() string {
	return "CreateAddressCommand"
}
