// internal/application/queries/list_addresses_query.go
package queries

// ListAddressesQuery representa la query para listar direcciones
type ListAddressesQuery struct {
	UserID string
}

func (q ListAddressesQuery) GetName() string {
	return "ListAddressesQuery"
}
