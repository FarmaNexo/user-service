// internal/application/preprocessors/sanitize_input_preprocessor.go
package preprocessors

import (
	"context"
	"strings"

	"github.com/farmanexo/user-service/internal/application/commands"
	"go.uber.org/zap"
)

// SanitizeInputPreProcessor limpia y normaliza los inputs
type SanitizeInputPreProcessor struct {
	logger *zap.Logger
}

func NewSanitizeInputPreProcessor(logger *zap.Logger) *SanitizeInputPreProcessor {
	return &SanitizeInputPreProcessor{logger: logger}
}

func (p *SanitizeInputPreProcessor) Process(ctx context.Context, request interface{}) error {
	switch cmd := request.(type) {
	case *commands.UpdateProfileCommand:
		cmd.FullName = strings.TrimSpace(cmd.FullName)
		cmd.FullName = p.titleCase(cmd.FullName)
		cmd.Phone = strings.TrimSpace(cmd.Phone)
		cmd.Bio = strings.TrimSpace(cmd.Bio)
		p.logger.Debug("Input sanitizado", zap.String("command", "UpdateProfileCommand"))

	case *commands.CreateAddressCommand:
		cmd.Label = strings.TrimSpace(cmd.Label)
		cmd.Street = strings.TrimSpace(cmd.Street)
		cmd.City = strings.TrimSpace(cmd.City)
		cmd.State = strings.TrimSpace(cmd.State)
		cmd.PostalCode = strings.TrimSpace(cmd.PostalCode)
		cmd.Country = strings.TrimSpace(cmd.Country)
		p.logger.Debug("Input sanitizado", zap.String("command", "CreateAddressCommand"))

	case *commands.UpdateAddressCommand:
		cmd.Label = strings.TrimSpace(cmd.Label)
		cmd.Street = strings.TrimSpace(cmd.Street)
		cmd.City = strings.TrimSpace(cmd.City)
		cmd.State = strings.TrimSpace(cmd.State)
		cmd.PostalCode = strings.TrimSpace(cmd.PostalCode)
		cmd.Country = strings.TrimSpace(cmd.Country)
		p.logger.Debug("Input sanitizado", zap.String("command", "UpdateAddressCommand"))
	}

	return nil
}

func (p *SanitizeInputPreProcessor) titleCase(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}
