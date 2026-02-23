// internal/application/queries/get_profile_query.go
package queries

// GetProfileQuery representa la query para obtener el perfil
type GetProfileQuery struct {
	UserID string
}

func (q GetProfileQuery) GetName() string {
	return "GetProfileQuery"
}
