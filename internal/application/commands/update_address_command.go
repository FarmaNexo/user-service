// internal/application/commands/update_address_command.go
package commands

// UpdateAddressCommand representa el comando para actualizar una dirección
type UpdateAddressCommand struct {
	UserID     string
	AddressID  string
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

func (c UpdateAddressCommand) GetName() string {
	return "UpdateAddressCommand"
}
