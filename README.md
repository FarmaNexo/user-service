# FarmaNexo User Service

Microservicio de gestión de perfil de usuario para FarmaNexo.

## Stack

- **Go 1.25+** con Chi Router
- **PostgreSQL** con GORM
- **AWS S3** (LocalStack local) para avatares
- **AWS SQS** para eventos asíncronos
- **JWT** para autenticación (validación, mismo secret que Auth Service)

## Inicio Rápido

```bash
# 1. Instalar dependencias
make install

# 2. Crear base de datos
psql -U admin -h localhost -c "CREATE DATABASE user_db;"

# 3. Ejecutar migraciones
make migrate-up

# 4. Ejecutar servicio
make dev
```

El servicio estará disponible en `http://localhost:4002`
Swagger UI: `http://localhost:4002/swagger/index.html`

## Endpoints

Todos los endpoints requieren header `Authorization: Bearer {token}`.

### Perfil

```bash
# Obtener perfil
curl -H "Authorization: Bearer TOKEN" http://localhost:4002/api/v1/users/me

# Actualizar perfil
curl -X PUT -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"full_name":"Juan Pérez","phone":"+51999888777","bio":"Farmacéutico"}' \
  http://localhost:4002/api/v1/users/me
```

### Avatar

```bash
# Subir avatar (max 5MB, jpg/png/webp)
curl -X PUT -H "Authorization: Bearer TOKEN" \
  -F "avatar=@foto.jpg" \
  http://localhost:4002/api/v1/users/me/avatar

# Eliminar avatar
curl -X DELETE -H "Authorization: Bearer TOKEN" \
  http://localhost:4002/api/v1/users/me/avatar
```

### Direcciones

```bash
# Listar direcciones
curl -H "Authorization: Bearer TOKEN" http://localhost:4002/api/v1/users/me/addresses

# Crear dirección
curl -X POST -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"label":"Casa","street":"Av. Arequipa 1234","city":"Lima","country":"Perú","is_default":true}' \
  http://localhost:4002/api/v1/users/me/addresses

# Actualizar dirección
curl -X PUT -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"label":"Oficina","street":"Av. Javier Prado 5678","city":"Lima","country":"Perú","is_default":false}' \
  http://localhost:4002/api/v1/users/me/addresses/{id}

# Eliminar dirección
curl -X DELETE -H "Authorization: Bearer TOKEN" \
  http://localhost:4002/api/v1/users/me/addresses/{id}
```

### Preferencias

```bash
# Obtener preferencias
curl -H "Authorization: Bearer TOKEN" http://localhost:4002/api/v1/users/me/preferences

# Actualizar preferencias
curl -X PUT -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"language":"es","theme":"dark","notifications_enabled":true,"email_notifications":true,"sms_notifications":false}' \
  http://localhost:4002/api/v1/users/me/preferences
```

## Arquitectura

```
cmd/server/main.go           → Entry point
internal/
  application/               → CQRS Commands, Queries, Handlers
  domain/                    → Entities, Repository interfaces
  infrastructure/            → PostgreSQL, S3, SQS, JWT
  presentation/              → HTTP Controllers, DTOs, Middlewares
  shared/                    → ApiResponse, Constants
pkg/                         → Config, Mediator, Logger
migrations/                  → SQL migrations
configs/                     → YAML per environment
```

## Eventos

### Consume (de Auth Service)
- `USER_REGISTERED` → Crea perfil inicial + preferencias por defecto

### Publica
- `PROFILE_UPDATED` → Cuando se actualiza el perfil
- `AVATAR_CHANGED` → Cuando se sube/elimina avatar
- `ADDRESS_CREATED` → Cuando se crea una dirección
