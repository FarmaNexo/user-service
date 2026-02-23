// internal/application/validators/update_profile_validator.go
package validators

import (
	"context"
	"fmt"
	"regexp"

	"github.com/farmanexo/user-service/internal/application/commands"
)

// UpdateProfileValidator valida el comando UpdateProfileCommand
type UpdateProfileValidator struct {
	phoneRegex *regexp.Regexp
}

func NewUpdateProfileValidator() *UpdateProfileValidator {
	return &UpdateProfileValidator{
		phoneRegex: regexp.MustCompile(`^[\+]?[0-9]{7,15}$`),
	}
}

func (v *UpdateProfileValidator) Validate(ctx context.Context, cmd interface{}) error {
	command, ok := cmd.(commands.UpdateProfileCommand)
	if !ok {
		return nil
	}

	if command.FullName == "" {
		return fmt.Errorf("el nombre completo es requerido")
	}

	if len(command.FullName) < 2 || len(command.FullName) > 255 {
		return fmt.Errorf("el nombre debe tener entre 2 y 255 caracteres")
	}

	if command.Phone != "" && !v.phoneRegex.MatchString(command.Phone) {
		return fmt.Errorf("formato de teléfono inválido")
	}

	if len(command.Bio) > 500 {
		return fmt.Errorf("la biografía no puede exceder 500 caracteres")
	}

	return nil
}
