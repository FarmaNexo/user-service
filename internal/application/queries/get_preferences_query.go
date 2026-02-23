// internal/application/queries/get_preferences_query.go
package queries

// GetPreferencesQuery representa la query para obtener preferencias
type GetPreferencesQuery struct {
	UserID string
}

func (q GetPreferencesQuery) GetName() string {
	return "GetPreferencesQuery"
}
