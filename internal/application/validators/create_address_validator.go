// internal/application/validators/create_address_validator.go
package validators

import (
	"context"
	"fmt"

	"github.com/farmanexo/user-service/internal/application/commands"
)

// CreateAddressValidator valida el comando CreateAddressCommand
type CreateAddressValidator struct{}

func NewCreateAddressValidator() *CreateAddressValidator {
	return &CreateAddressValidator{}
}

func (v *CreateAddressValidator) Validate(ctx context.Context, cmd interface{}) error {
	command, ok := cmd.(commands.CreateAddressCommand)
	if !ok {
		return nil
	}

	if command.Street == "" {
		return fmt.Errorf("la calle es requerida")
	}

	if command.City == "" {
		return fmt.Errorf("la ciudad es requerida")
	}

	if len(command.Label) > 50 {
		return fmt.Errorf("la etiqueta no puede exceder 50 caracteres")
	}

	if len(command.Street) > 255 {
		return fmt.Errorf("la calle no puede exceder 255 caracteres")
	}

	return nil
}
