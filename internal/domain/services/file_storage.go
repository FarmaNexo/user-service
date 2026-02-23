// internal/domain/services/file_storage.go
package services

import (
	"context"
	"io"
)

// FileStorage define la interfaz para almacenamiento de archivos
type FileStorage interface {
	// Upload sube un archivo y retorna la URL pública
	Upload(ctx context.Context, bucket, key string, reader io.Reader, contentType string) (string, error)

	// Delete elimina un archivo del storage
	Delete(ctx context.Context, bucket, key string) error
}
