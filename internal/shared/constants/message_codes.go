// internal/shared/constants/message_codes.go
package constants

// MessageCode contiene todos los códigos de respuesta del sistema
type MessageCode string

const (
	// Success codes
	CodeSuccess        MessageCode = "SUCCESS_001"
	CodeCreatedSuccess MessageCode = "SUCCESS_002"
	CodeUpdatedSuccess MessageCode = "SUCCESS_003"
	CodeDeletedSuccess MessageCode = "SUCCESS_004"

	// User profile codes
	CodeProfileRetrieved  MessageCode = "USR_001"
	CodeProfileUpdated    MessageCode = "USR_002"
	CodeAvatarUploaded    MessageCode = "USR_003"
	CodeAvatarDeleted     MessageCode = "USR_004"
	CodeAddressCreated    MessageCode = "USR_005"
	CodeAddressUpdated    MessageCode = "USR_006"
	CodeAddressDeleted    MessageCode = "USR_007"
	CodePreferencesLoaded MessageCode = "USR_008"
	CodePreferencesUpdate MessageCode = "USR_009"

	// Validation errors
	CodeValidationError  MessageCode = "VAL_001"
	CodeInvalidEmail     MessageCode = "VAL_002"
	CodeInvalidPhone     MessageCode = "VAL_003"
	CodeInvalidFileType  MessageCode = "VAL_004"
	CodeFileTooLarge     MessageCode = "VAL_005"
	CodeRequiredField    MessageCode = "VAL_006"
	CodeInvalidLanguage  MessageCode = "VAL_007"
	CodeInvalidTheme     MessageCode = "VAL_008"

	// Authentication errors
	CodeUnauthorized       MessageCode = "AUTH_ERR_001"
	CodeInvalidToken       MessageCode = "AUTH_ERR_002"
	CodeTokenExpired       MessageCode = "AUTH_ERR_003"
	CodeInvalidCredentials MessageCode = "AUTH_ERR_004"

	// Business errors
	CodeProfileNotFound  MessageCode = "BUS_001"
	CodeAddressNotFound  MessageCode = "BUS_002"
	CodeResourceNotFound MessageCode = "BUS_003"
	CodeMaxAddresses     MessageCode = "BUS_004"

	// Rate limiting
	CodeRateLimitExceeded MessageCode = "RATE_001"

	// System errors
	CodeInternalError      MessageCode = "SYS_001"
	CodeDatabaseError      MessageCode = "SYS_002"
	CodeServiceUnavailable MessageCode = "SYS_003"
	CodeStorageError       MessageCode = "SYS_004"
)

// MessageDescription contiene las descripciones predefinidas
var MessageDescription = map[MessageCode]string{
	// Success
	CodeSuccess:        "Operación exitosa",
	CodeCreatedSuccess: "Recurso creado exitosamente",
	CodeUpdatedSuccess: "Recurso actualizado exitosamente",
	CodeDeletedSuccess: "Recurso eliminado exitosamente",

	// User profile
	CodeProfileRetrieved:  "Perfil obtenido exitosamente",
	CodeProfileUpdated:    "Perfil actualizado exitosamente",
	CodeAvatarUploaded:    "Avatar subido exitosamente",
	CodeAvatarDeleted:     "Avatar eliminado exitosamente",
	CodeAddressCreated:    "Dirección creada exitosamente",
	CodeAddressUpdated:    "Dirección actualizada exitosamente",
	CodeAddressDeleted:    "Dirección eliminada exitosamente",
	CodePreferencesLoaded: "Preferencias obtenidas exitosamente",
	CodePreferencesUpdate: "Preferencias actualizadas exitosamente",

	// Validation
	CodeValidationError: "Error de validación",
	CodeInvalidEmail:    "El formato del email es inválido",
	CodeInvalidPhone:    "El formato del teléfono es inválido",
	CodeInvalidFileType: "Tipo de archivo no permitido. Use: jpg, jpeg, png, webp",
	CodeFileTooLarge:    "El archivo excede el tamaño máximo de 5MB",
	CodeRequiredField:   "Campo requerido",
	CodeInvalidLanguage: "Idioma no válido. Use: es, en",
	CodeInvalidTheme:    "Tema no válido. Use: light, dark",

	// Auth errors
	CodeUnauthorized:       "No autorizado",
	CodeInvalidToken:       "Token inválido",
	CodeTokenExpired:       "Token expirado",
	CodeInvalidCredentials: "Credenciales inválidas",

	// Business
	CodeProfileNotFound:  "Perfil de usuario no encontrado",
	CodeAddressNotFound:  "Dirección no encontrada",
	CodeResourceNotFound: "Recurso no encontrado",
	CodeMaxAddresses:     "Máximo de direcciones alcanzado",

	// Rate limiting
	CodeRateLimitExceeded: "Demasiadas solicitudes. Intente nuevamente más tarde",

	// System
	CodeInternalError:      "Error interno del servidor",
	CodeDatabaseError:      "Error de base de datos",
	CodeServiceUnavailable: "Servicio no disponible",
	CodeStorageError:       "Error en servicio de almacenamiento",
}

// GetDescription retorna la descripción del código
func GetDescription(code MessageCode) string {
	if desc, ok := MessageDescription[code]; ok {
		return desc
	}
	return "Descripción no disponible"
}
