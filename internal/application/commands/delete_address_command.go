// internal/application/commands/delete_address_command.go
package commands

// DeleteAddressCommand representa el comando para eliminar una dirección
type DeleteAddressCommand struct {
	UserID    string
	AddressID string
}

func (c DeleteAddressCommand) GetName() string {
	return "DeleteAddressCommand"
}
