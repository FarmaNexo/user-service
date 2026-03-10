# User Service

Microservicio de gestion de perfil de usuario para FarmaNexo. Maneja perfiles, avatares (S3), direcciones y preferencias. Consume eventos `USER_REGISTERED` de Auth Service para crear perfiles iniciales automaticamente.

**NO** genera tokens JWT, solo los valida usando el mismo secret que Auth Service.

## Inicio Rapido

### Prerequisitos
- Go 1.25+
- PostgreSQL 16
- Redis 7
- LocalStack (desarrollo local)
- Docker & Docker Compose

### Instalacion
```bash
# Clonar repositorio
git clone <url>
cd services/user-service

# Instalar dependencias
go mod download

# Configurar ambiente local
cp configs/config.development.yaml configs/config.local.yaml
# Editar configs/config.local.yaml con tus credenciales

# Crear base de datos
docker exec -it farmanexo-postgres psql -U admin -c "CREATE DATABASE user_db;"

# Ejecutar migraciones
make migrate-up

# Ejecutar servicio
make dev
```

Swagger UI disponible en: http://localhost:4002/swagger/index.html

## Endpoints

Todos los endpoints (excepto health) requieren `Authorization: Bearer {token}`.

### Perfil

**GET /api/v1/users/me** - Obtener perfil
```bash
curl -H "Authorization: Bearer TOKEN" \
  http://localhost:4002/api/v1/users/me
```

Respuesta (200):
```json
{
  "meta": {
    "mensajes": [{ "codigo": "USR_001", "mensaje": "Perfil obtenido exitosamente", "tipo": "exito" }],
    "idTransaccion": "...",
    "resultado": true,
    "timestamp": "20260222 103000"
  },
  "datos": {
    "user_id": "uuid",
    "full_name": "Juan Perez",
    "phone": "+51999888777",
    "bio": "Farmaceutico",
    "avatar_url": "http://localhost:4566/farmanexo-avatars/users/uuid/avatar.jpg",
    "date_of_birth": "1990-01-15",
    "created_at": "2026-02-22T10:30:00Z",
    "updated_at": "2026-02-22T10:30:00Z"
  }
}
```

**PUT /api/v1/users/me** - Actualizar perfil
```bash
curl -X PUT -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Juan Perez",
    "phone": "+51999888777",
    "bio": "Farmaceutico profesional",
    "date_of_birth": "1990-01-15"
  }' \
  http://localhost:4002/api/v1/users/me
```

### Avatar

**PUT /api/v1/users/me/avatar** - Subir avatar
```bash
curl -X PUT -H "Authorization: Bearer TOKEN" \
  -F "avatar=@foto.jpg" \
  http://localhost:4002/api/v1/users/me/avatar
```

- Formatos: jpg, jpeg, png, webp
- Tamano maximo: 5 MB
- Se redimensiona automaticamente a 512x512px
- Calidad JPEG: 85%

**DELETE /api/v1/users/me/avatar** - Eliminar avatar
```bash
curl -X DELETE -H "Authorization: Bearer TOKEN" \
  http://localhost:4002/api/v1/users/me/avatar
```

### Direcciones

**GET /api/v1/users/me/addresses** - Listar direcciones
```bash
curl -H "Authorization: Bearer TOKEN" \
  http://localhost:4002/api/v1/users/me/addresses
```

**POST /api/v1/users/me/addresses** - Crear direccion
```bash
curl -X POST -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "label": "Casa",
    "street": "Av. Arequipa 1234",
    "city": "Lima",
    "state": "Lima",
    "postal_code": "15001",
    "country": "Peru",
    "is_default": true,
    "latitude": -12.0464,
    "longitude": -77.0428
  }' \
  http://localhost:4002/api/v1/users/me/addresses
```

**PUT /api/v1/users/me/addresses/{id}** - Actualizar direccion
```bash
curl -X PUT -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "label": "Oficina",
    "street": "Av. Javier Prado 5678",
    "city": "Lima",
    "country": "Peru",
    "is_default": false
  }' \
  http://localhost:4002/api/v1/users/me/addresses/{id}
```

**DELETE /api/v1/users/me/addresses/{id}** - Eliminar direccion
```bash
curl -X DELETE -H "Authorization: Bearer TOKEN" \
  http://localhost:4002/api/v1/users/me/addresses/{id}
```

### Preferencias

**GET /api/v1/users/me/preferences** - Obtener preferencias
```bash
curl -H "Authorization: Bearer TOKEN" \
  http://localhost:4002/api/v1/users/me/preferences
```

**PUT /api/v1/users/me/preferences** - Actualizar preferencias
```bash
curl -X PUT -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "language": "es",
    "theme": "dark",
    "notifications_enabled": true,
    "email_notifications": true,
    "sms_notifications": false
  }' \
  http://localhost:4002/api/v1/users/me/preferences
```

### Health Check

**GET /health** - Estado del servicio
```bash
curl http://localhost:4002/health
```

## Arquitectura

- **Puerto:** 4002
- **Base de datos:** `user_db`
- **Schema:** `user`
- **Patron:** Clean Architecture + CQRS + MediatR

### Capas

```
internal/
  domain/           Entidades (UserProfile, Address, Preferences), interfaces
  application/      Commands (8), Queries (3), Handlers (9+), Validators, Pre/Post processors
  infrastructure/   PostgreSQL (GORM), S3, SQS publisher/consumer, JWT
  presentation/     Controllers, Middlewares, Routes, DTOs
  shared/           ApiResponse[T], Constants, Errors
pkg/
  mediator/         Mediator CQRS generico con pipeline
  config/           Carga de configuracion por ambiente (Viper)
```

### Flujo de un request

```
HTTP Request
  -> Chi Router
    -> [Middlewares: RequestID, RealIP, Logger, Recoverer, CORS, CorrelationID]
    -> AuthMiddleware (valida JWT)
    -> Controller
      -> Mediator.Send(Command/Query)
        -> Validator
        -> PreProcessor (SanitizeInput)
        -> Handler
        -> PostProcessor (LogAudit)
      <- ApiResponse[T]
    <- JSON Response
```

## Configuracion

### Variables de Entorno

```yaml
# configs/config.local.yaml
environment: local

server:
  host: 0.0.0.0
  port: 4002
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s

database:
  host: localhost
  port: 5432
  user: admin
  password: admin
  db_name: user_db
  schema: "user"
  sslmode: disable
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

jwt:
  secret: "dev-super-secret-key-change-in-production-min-32-chars"
  access_token_duration: 15m
  issuer: "farmanexo-user-service"

aws:
  region: us-east-1
  endpoint: "http://localhost:4566"

sqs:
  user_events_queue_url: "http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/farmanexo-user-events"
  auth_events_queue_url: "http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/farmanexo-auth-events"

s3:
  avatars_bucket: "farmanexo-avatars"

log:
  level: debug
  encoding: console
```

### Variables de entorno requeridas (ambientes desplegados)

| Variable | Descripcion |
|---|---|
| `DB_HOST` | Host de PostgreSQL |
| `DB_USER` | Usuario de PostgreSQL |
| `DB_PASSWORD` | Password de PostgreSQL |
| `JWT_SECRET` | Secret para validar JWT (mismo que Auth Service) |
| `REDIS_HOST` | Host de Redis |
| `REDIS_PASSWORD` | Password de Redis |
| `AWS_REGION` | Region de AWS |
| `SQS_USER_EVENTS_QUEUE_URL` | URL cola para publicar eventos |
| `SQS_AUTH_EVENTS_QUEUE_URL` | URL cola para consumir eventos de Auth |

## Infraestructura

### PostgreSQL
- **Database:** `user_db`
- **Schema:** `user`
- **Tablas:** `profiles`, `addresses`, `preferences`

### Redis
- No utilizado directamente por este servicio (planificado para cache futuro)

### S3
- **Bucket:** `farmanexo-avatars`
- **Path:** `users/{user_id}/avatar.{ext}`
- **Procesamiento:** Resize a 512x512, calidad 85%

### SQS
- **Consume:** `farmanexo-auth-events` (evento `USER_REGISTERED`)
- **Publica en:** `farmanexo-user-events` (eventos `PROFILE_UPDATED`, `AVATAR_CHANGED`, `ADDRESS_CREATED`)

## Eventos

### Consume (de Auth Service)

| Evento | Cola | Accion |
|---|---|---|
| `USER_REGISTERED` | `farmanexo-auth-events` | Crea perfil inicial + preferencias por defecto |

- Long polling: 20s wait, max 10 mensajes por receive
- Idempotente: verifica si el perfil ya existe antes de crear

### Publica

| Evento | Cola | Trigger |
|---|---|---|
| `PROFILE_UPDATED` | `farmanexo-user-events` | Actualizacion de perfil |
| `AVATAR_CHANGED` | `farmanexo-user-events` | Upload o eliminacion de avatar |
| `ADDRESS_CREATED` | `farmanexo-user-events` | Creacion de direccion |

Formato:
```json
{
  "event_type": "PROFILE_UPDATED",
  "user_id": "uuid",
  "timestamp": "2026-02-22T12:00:00Z",
  "metadata": {
    "source": "user-service",
    "version": "1.0"
  }
}
```

## Testing
```bash
# Unit tests
make test

# Tests con coverage
make test-coverage

# Generar mocks
make gen-mocks
```

## Comandos Utiles
```bash
# Desarrollo
make dev              # Ejecutar en modo desarrollo
make build            # Compilar binario a bin/user-service
make swagger          # Generar documentacion Swagger

# Base de datos
make migrate-up       # Aplicar migraciones pendientes
make migrate-down     # Revertir ultima migracion
make migrate-create NAME=nombre  # Crear nueva migracion

# Calidad
make lint             # Ejecutar golangci-lint
make format           # Formatear codigo con goimports

# Docker
make docker-build     # Construir imagen Docker
make docker-run       # Ejecutar container
```

## Dependencias

### Principales
- `github.com/go-chi/chi/v5` - HTTP router
- `gorm.io/gorm` - ORM
- `github.com/aws/aws-sdk-go-v2` - AWS SDK (S3, SQS)
- `github.com/golang-jwt/jwt/v5` - JWT validation
- `github.com/disintegration/imaging` - Procesamiento de imagenes
- `go.uber.org/zap` - Structured logging
- `github.com/spf13/viper` - Configuracion
- `github.com/swaggo/swag` - Swagger

### Completas
Ver `go.mod`

## Documentacion Adicional

- [CLAUDE.md](./CLAUDE.md) - Contexto para Claude AI
- [INFRASTRUCTURE.md](./INFRASTRUCTURE.md) - Detalle de infraestructura
- [Swagger UI](http://localhost:4002/swagger/index.html) - API docs interactiva

## Estructura de Directorios

```
user-service/
  cmd/server/main.go                    Punto de entrada con DI
  configs/                              YAML por ambiente (5 archivos)
  migrations/                           SQL (golang-migrate, schema: user)
  internal/
    application/
      commands/                         UpdateProfile, UploadAvatar, DeleteAvatar,
                                        CreateAddress, UpdateAddress, DeleteAddress,
                                        UpdatePreferences
      queries/                          GetProfile, ListAddresses, GetPreferences
      handlers/                         Handler por cada command/query (9+)
      validators/                       Validadores de commands
      preprocessors/                    SanitizeInput
      postprocessors/                   LogAudit
    domain/
      entities/                         UserProfile, Address, Preferences
      events/                           Eventos de usuario y auth
      repositories/                     ProfileRepository, AddressRepository, PreferencesRepository
      services/                         FileStorage, EventPublisher (interfaces)
    infrastructure/
      persistence/postgres/             Repositorios GORM (3)
      messaging/                        SQS publisher + consumer
      storage/                          S3 file storage
      security/                         JWT service (validacion solamente)
    presentation/
      dto/requests/                     UpdateProfileRequest, CreateAddressRequest, etc.
      dto/responses/                    ProfileResponse, AddressResponse, etc.
      http/controllers/                 UserController
      http/middlewares/                 AuthMiddleware, CorrelationID
      http/routes/                      Configuracion de rutas Chi
    shared/
      common/                           ApiResponse[T], response factories
      constants/                        Codigos HTTP, message codes (55+)
      errors/                           Domain errors
  pkg/
    config/                             Carga de configuracion (Viper)
    mediator/                           Mediator CQRS generico
  docs/                                 Swagger generado
```
