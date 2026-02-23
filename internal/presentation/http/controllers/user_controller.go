// internal/presentation/http/controllers/user_controller.go
package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/farmanexo/user-service/internal/application/commands"
	"github.com/farmanexo/user-service/internal/application/queries"
	"github.com/farmanexo/user-service/internal/presentation/dto/requests"
	"github.com/farmanexo/user-service/internal/presentation/dto/responses"
	"github.com/farmanexo/user-service/internal/presentation/http/middlewares"
	"github.com/farmanexo/user-service/internal/shared/common"
	"github.com/farmanexo/user-service/internal/shared/constants"
	"github.com/farmanexo/user-service/pkg/mediator"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// UserController maneja los endpoints de usuario
type UserController struct {
	mediator *mediator.Mediator
	logger   *zap.Logger
}

func NewUserController(mediator *mediator.Mediator, logger *zap.Logger) *UserController {
	return &UserController{
		mediator: mediator,
		logger:   logger,
	}
}

// ========================================
// PROFILE ENDPOINTS
// ========================================

// GetProfile godoc
// @Summary      Obtener perfil del usuario
// @Description  Retorna el perfil completo del usuario autenticado
// @Tags         Profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  common.ApiResponse[responses.ProfileResponse]  "Perfil obtenido exitosamente"
// @Failure      401  {object}  common.ApiResponse[responses.ProfileResponse]  "No autorizado"
// @Failure      404  {object}  common.ApiResponse[responses.ProfileResponse]  "Perfil no encontrado"
// @Failure      500  {object}  common.ApiResponse[responses.ProfileResponse]  "Error interno"
// @Router       /api/v1/users/me [get]
func (c *UserController) GetProfile(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("GET /api/v1/users/me")

	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		c.respondJSON(w, common.UnauthorizedResponse[responses.ProfileResponse]("Usuario no autenticado"))
		return
	}

	query := queries.GetProfileQuery{UserID: userID}
	response, err := mediator.Send[queries.GetProfileQuery, responses.ProfileResponse](
		r.Context(), c.mediator, query,
	)
	if err != nil {
		c.logger.Error("Error ejecutando GetProfileQuery", zap.Error(err))
		c.respondJSON(w, common.InternalServerErrorResponse[responses.ProfileResponse]("Error obteniendo perfil"))
		return
	}

	c.respondJSON(w, response)
}

// UpdateProfile godoc
// @Summary      Actualizar perfil del usuario
// @Description  Actualiza los datos del perfil del usuario autenticado
// @Tags         Profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      requests.UpdateProfileRequest  true  "Datos a actualizar"
// @Success      200      {object}  common.ApiResponse[responses.ProfileResponse]  "Perfil actualizado"
// @Failure      400      {object}  common.ApiResponse[responses.ProfileResponse]  "Error de validación"
// @Failure      401      {object}  common.ApiResponse[responses.ProfileResponse]  "No autorizado"
// @Failure      500      {object}  common.ApiResponse[responses.ProfileResponse]  "Error interno"
// @Router       /api/v1/users/me [put]
func (c *UserController) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("PUT /api/v1/users/me")

	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		c.respondJSON(w, common.UnauthorizedResponse[responses.ProfileResponse]("Usuario no autenticado"))
		return
	}

	var req requests.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.ProfileResponse](
			constants.CodeValidationError, "Cuerpo de solicitud inválido",
		))
		return
	}

	command := commands.UpdateProfileCommand{
		UserID:      userID,
		FullName:    req.FullName,
		Phone:       req.Phone,
		Bio:         req.Bio,
		DateOfBirth: req.DateOfBirth,
	}

	response, err := mediator.Send[commands.UpdateProfileCommand, responses.ProfileResponse](
		r.Context(), c.mediator, command,
	)
	if err != nil {
		c.logger.Error("Error ejecutando UpdateProfileCommand", zap.Error(err))
		c.respondJSON(w, common.InternalServerErrorResponse[responses.ProfileResponse]("Error actualizando perfil"))
		return
	}

	c.respondJSON(w, response)
}

// ========================================
// AVATAR ENDPOINTS
// ========================================

// UploadAvatar godoc
// @Summary      Subir avatar del usuario
// @Description  Sube una imagen de avatar (max 5MB, formatos: jpg, png, webp)
// @Tags         Avatar
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        avatar  formData  file  true  "Archivo de imagen (max 5MB)"
// @Success      200     {object}  common.ApiResponse[responses.AvatarResponse]  "Avatar subido"
// @Failure      400     {object}  common.ApiResponse[responses.AvatarResponse]  "Archivo inválido"
// @Failure      401     {object}  common.ApiResponse[responses.AvatarResponse]  "No autorizado"
// @Failure      500     {object}  common.ApiResponse[responses.AvatarResponse]  "Error interno"
// @Router       /api/v1/users/me/avatar [put]
func (c *UserController) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("PUT /api/v1/users/me/avatar")

	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		c.respondJSON(w, common.UnauthorizedResponse[responses.AvatarResponse]("Usuario no autenticado"))
		return
	}

	// Limitar tamaño del body a 6MB (5MB + overhead del multipart)
	r.Body = http.MaxBytesReader(w, r.Body, 6*1024*1024)

	if err := r.ParseMultipartForm(5 * 1024 * 1024); err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.AvatarResponse](
			constants.CodeFileTooLarge, "El archivo excede el tamaño máximo de 5MB",
		))
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.AvatarResponse](
			constants.CodeRequiredField, "El campo 'avatar' es requerido",
		))
		return
	}
	defer file.Close()

	command := commands.UploadAvatarCommand{
		UserID:      userID,
		File:        file,
		Filename:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		FileSize:    header.Size,
	}

	response, err := mediator.Send[commands.UploadAvatarCommand, responses.AvatarResponse](
		r.Context(), c.mediator, command,
	)
	if err != nil {
		c.logger.Error("Error ejecutando UploadAvatarCommand", zap.Error(err))
		c.respondJSON(w, common.InternalServerErrorResponse[responses.AvatarResponse]("Error subiendo avatar"))
		return
	}

	c.respondJSON(w, response)
}

// DeleteAvatar godoc
// @Summary      Eliminar avatar del usuario
// @Description  Elimina el avatar del usuario autenticado
// @Tags         Avatar
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  common.ApiResponse[responses.EmptyResponse]  "Avatar eliminado"
// @Failure      401  {object}  common.ApiResponse[responses.EmptyResponse]  "No autorizado"
// @Failure      500  {object}  common.ApiResponse[responses.EmptyResponse]  "Error interno"
// @Router       /api/v1/users/me/avatar [delete]
func (c *UserController) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("DELETE /api/v1/users/me/avatar")

	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		c.respondJSON(w, common.UnauthorizedResponse[responses.EmptyResponse]("Usuario no autenticado"))
		return
	}

	command := commands.DeleteAvatarCommand{UserID: userID}
	response, err := mediator.Send[commands.DeleteAvatarCommand, responses.EmptyResponse](
		r.Context(), c.mediator, command,
	)
	if err != nil {
		c.logger.Error("Error ejecutando DeleteAvatarCommand", zap.Error(err))
		c.respondJSON(w, common.InternalServerErrorResponse[responses.EmptyResponse]("Error eliminando avatar"))
		return
	}

	c.respondJSON(w, response)
}

// ========================================
// ADDRESS ENDPOINTS
// ========================================

// ListAddresses godoc
// @Summary      Listar direcciones del usuario
// @Description  Retorna todas las direcciones del usuario autenticado
// @Tags         Addresses
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  common.ApiResponse[responses.AddressListResponse]  "Lista de direcciones"
// @Failure      401  {object}  common.ApiResponse[responses.AddressListResponse]  "No autorizado"
// @Failure      500  {object}  common.ApiResponse[responses.AddressListResponse]  "Error interno"
// @Router       /api/v1/users/me/addresses [get]
func (c *UserController) ListAddresses(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("GET /api/v1/users/me/addresses")

	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		c.respondJSON(w, common.UnauthorizedResponse[responses.AddressListResponse]("Usuario no autenticado"))
		return
	}

	query := queries.ListAddressesQuery{UserID: userID}
	response, err := mediator.Send[queries.ListAddressesQuery, responses.AddressListResponse](
		r.Context(), c.mediator, query,
	)
	if err != nil {
		c.logger.Error("Error ejecutando ListAddressesQuery", zap.Error(err))
		c.respondJSON(w, common.InternalServerErrorResponse[responses.AddressListResponse]("Error listando direcciones"))
		return
	}

	c.respondJSON(w, response)
}

// CreateAddress godoc
// @Summary      Crear una dirección
// @Description  Crea una nueva dirección para el usuario autenticado
// @Tags         Addresses
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      requests.CreateAddressRequest  true  "Datos de la dirección"
// @Success      201      {object}  common.ApiResponse[responses.AddressResponse]  "Dirección creada"
// @Failure      400      {object}  common.ApiResponse[responses.AddressResponse]  "Error de validación"
// @Failure      401      {object}  common.ApiResponse[responses.AddressResponse]  "No autorizado"
// @Failure      500      {object}  common.ApiResponse[responses.AddressResponse]  "Error interno"
// @Router       /api/v1/users/me/addresses [post]
func (c *UserController) CreateAddress(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("POST /api/v1/users/me/addresses")

	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		c.respondJSON(w, common.UnauthorizedResponse[responses.AddressResponse]("Usuario no autenticado"))
		return
	}

	var req requests.CreateAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.AddressResponse](
			constants.CodeValidationError, "Cuerpo de solicitud inválido",
		))
		return
	}

	command := commands.CreateAddressCommand{
		UserID:     userID,
		Label:      req.Label,
		Street:     req.Street,
		City:       req.City,
		State:      req.State,
		PostalCode: req.PostalCode,
		Country:    req.Country,
		IsDefault:  req.IsDefault,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
	}

	response, err := mediator.Send[commands.CreateAddressCommand, responses.AddressResponse](
		r.Context(), c.mediator, command,
	)
	if err != nil {
		c.logger.Error("Error ejecutando CreateAddressCommand", zap.Error(err))
		c.respondJSON(w, common.InternalServerErrorResponse[responses.AddressResponse]("Error creando dirección"))
		return
	}

	c.respondJSON(w, response)
}

// UpdateAddress godoc
// @Summary      Actualizar una dirección
// @Description  Actualiza una dirección existente del usuario autenticado
// @Tags         Addresses
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                          true  "ID de la dirección"
// @Param        request  body      requests.UpdateAddressRequest   true  "Datos a actualizar"
// @Success      200      {object}  common.ApiResponse[responses.AddressResponse]  "Dirección actualizada"
// @Failure      400      {object}  common.ApiResponse[responses.AddressResponse]  "Error de validación"
// @Failure      401      {object}  common.ApiResponse[responses.AddressResponse]  "No autorizado"
// @Failure      404      {object}  common.ApiResponse[responses.AddressResponse]  "Dirección no encontrada"
// @Failure      500      {object}  common.ApiResponse[responses.AddressResponse]  "Error interno"
// @Router       /api/v1/users/me/addresses/{id} [put]
func (c *UserController) UpdateAddress(w http.ResponseWriter, r *http.Request) {
	addressID := chi.URLParam(r, "id")
	c.logger.Info("PUT /api/v1/users/me/addresses/"+addressID)

	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		c.respondJSON(w, common.UnauthorizedResponse[responses.AddressResponse]("Usuario no autenticado"))
		return
	}

	var req requests.UpdateAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.AddressResponse](
			constants.CodeValidationError, "Cuerpo de solicitud inválido",
		))
		return
	}

	command := commands.UpdateAddressCommand{
		UserID:     userID,
		AddressID:  addressID,
		Label:      req.Label,
		Street:     req.Street,
		City:       req.City,
		State:      req.State,
		PostalCode: req.PostalCode,
		Country:    req.Country,
		IsDefault:  req.IsDefault,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
	}

	response, err := mediator.Send[commands.UpdateAddressCommand, responses.AddressResponse](
		r.Context(), c.mediator, command,
	)
	if err != nil {
		c.logger.Error("Error ejecutando UpdateAddressCommand", zap.Error(err))
		c.respondJSON(w, common.InternalServerErrorResponse[responses.AddressResponse]("Error actualizando dirección"))
		return
	}

	c.respondJSON(w, response)
}

// DeleteAddress godoc
// @Summary      Eliminar una dirección
// @Description  Elimina una dirección del usuario autenticado
// @Tags         Addresses
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string  true  "ID de la dirección"
// @Success      200  {object}  common.ApiResponse[responses.EmptyResponse]  "Dirección eliminada"
// @Failure      401  {object}  common.ApiResponse[responses.EmptyResponse]  "No autorizado"
// @Failure      404  {object}  common.ApiResponse[responses.EmptyResponse]  "Dirección no encontrada"
// @Failure      500  {object}  common.ApiResponse[responses.EmptyResponse]  "Error interno"
// @Router       /api/v1/users/me/addresses/{id} [delete]
func (c *UserController) DeleteAddress(w http.ResponseWriter, r *http.Request) {
	addressID := chi.URLParam(r, "id")
	c.logger.Info("DELETE /api/v1/users/me/addresses/"+addressID)

	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		c.respondJSON(w, common.UnauthorizedResponse[responses.EmptyResponse]("Usuario no autenticado"))
		return
	}

	command := commands.DeleteAddressCommand{UserID: userID, AddressID: addressID}
	response, err := mediator.Send[commands.DeleteAddressCommand, responses.EmptyResponse](
		r.Context(), c.mediator, command,
	)
	if err != nil {
		c.logger.Error("Error ejecutando DeleteAddressCommand", zap.Error(err))
		c.respondJSON(w, common.InternalServerErrorResponse[responses.EmptyResponse]("Error eliminando dirección"))
		return
	}

	c.respondJSON(w, response)
}

// ========================================
// PREFERENCES ENDPOINTS
// ========================================

// GetPreferences godoc
// @Summary      Obtener preferencias del usuario
// @Description  Retorna las preferencias del usuario autenticado
// @Tags         Preferences
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  common.ApiResponse[responses.PreferencesResponse]  "Preferencias obtenidas"
// @Failure      401  {object}  common.ApiResponse[responses.PreferencesResponse]  "No autorizado"
// @Failure      500  {object}  common.ApiResponse[responses.PreferencesResponse]  "Error interno"
// @Router       /api/v1/users/me/preferences [get]
func (c *UserController) GetPreferences(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("GET /api/v1/users/me/preferences")

	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		c.respondJSON(w, common.UnauthorizedResponse[responses.PreferencesResponse]("Usuario no autenticado"))
		return
	}

	query := queries.GetPreferencesQuery{UserID: userID}
	response, err := mediator.Send[queries.GetPreferencesQuery, responses.PreferencesResponse](
		r.Context(), c.mediator, query,
	)
	if err != nil {
		c.logger.Error("Error ejecutando GetPreferencesQuery", zap.Error(err))
		c.respondJSON(w, common.InternalServerErrorResponse[responses.PreferencesResponse]("Error obteniendo preferencias"))
		return
	}

	c.respondJSON(w, response)
}

// UpdatePreferences godoc
// @Summary      Actualizar preferencias del usuario
// @Description  Actualiza las preferencias del usuario autenticado
// @Tags         Preferences
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      requests.UpdatePreferencesRequest  true  "Preferencias a actualizar"
// @Success      200      {object}  common.ApiResponse[responses.PreferencesResponse]  "Preferencias actualizadas"
// @Failure      400      {object}  common.ApiResponse[responses.PreferencesResponse]  "Error de validación"
// @Failure      401      {object}  common.ApiResponse[responses.PreferencesResponse]  "No autorizado"
// @Failure      500      {object}  common.ApiResponse[responses.PreferencesResponse]  "Error interno"
// @Router       /api/v1/users/me/preferences [put]
func (c *UserController) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("PUT /api/v1/users/me/preferences")

	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		c.respondJSON(w, common.UnauthorizedResponse[responses.PreferencesResponse]("Usuario no autenticado"))
		return
	}

	var req requests.UpdatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.PreferencesResponse](
			constants.CodeValidationError, "Cuerpo de solicitud inválido",
		))
		return
	}

	command := commands.UpdatePreferencesCommand{
		UserID:               userID,
		Language:             req.Language,
		Theme:                req.Theme,
		NotificationsEnabled: req.NotificationsEnabled,
		EmailNotifications:   req.EmailNotifications,
		SMSNotifications:     req.SMSNotifications,
	}

	response, err := mediator.Send[commands.UpdatePreferencesCommand, responses.PreferencesResponse](
		r.Context(), c.mediator, command,
	)
	if err != nil {
		c.logger.Error("Error ejecutando UpdatePreferencesCommand", zap.Error(err))
		c.respondJSON(w, common.InternalServerErrorResponse[responses.PreferencesResponse]("Error actualizando preferencias"))
		return
	}

	c.respondJSON(w, response)
}

// ========================================
// HEALTH CHECK
// ========================================

// HealthCheck godoc
// @Summary      Health check
// @Description  Verifica el estado del servicio
// @Tags         Health
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Servicio saludable"
// @Router       /health [get]
func (c *UserController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	type HealthResponse struct {
		Status  string `json:"status" example:"healthy"`
		Service string `json:"service" example:"user-service"`
		Version string `json:"version" example:"1.0.0"`
	}

	health := HealthResponse{
		Status:  "healthy",
		Service: "user-service",
		Version: "1.0.0",
	}

	c.respondJSON(w, common.OkResponse(health))
}

// ========================================
// RESPONSE HELPER
// ========================================

func (c *UserController) respondJSON(w http.ResponseWriter, response interface{}) {
	statusCode := http.StatusOK

	if resp, ok := response.(interface{ GetHttpStatus() *int }); ok {
		if httpStatus := resp.GetHttpStatus(); httpStatus != nil {
			statusCode = *httpStatus
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		c.logger.Error("Error codificando respuesta JSON", zap.Error(err))
	}
}
