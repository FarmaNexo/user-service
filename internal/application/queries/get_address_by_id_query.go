// internal/application/queries/get_address_by_id_query.go
package queries

// GetAddressByIDQuery — lookup de una dirección por ID, restringido al
// usuario autenticado (ownership check en el handler). Lo consume
// order-service al validar la dirección de entrega en checkout.
type GetAddressByIDQuery struct {
	UserID    string
	AddressID string
}

func (q GetAddressByIDQuery) GetName() string {
	return "GetAddressByIDQuery"
}
