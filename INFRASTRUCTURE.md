# INFRAESTRUCTURA - User Service

## Resumen

Este documento describe la infraestructura especifica utilizada por User Service (puerto 4002).

---

## SERVICIOS REQUERIDOS

### PostgreSQL
- **Host:** localhost:5432 (local) / RDS endpoint (cloud)
- **Database:** `user_db`
- **User:** admin (local) / `${DB_USER}` (cloud)
- **Password:** admin (local) / `${DB_PASSWORD}` (cloud)
- **Schema:** `user`
- **Extensiones:** `gen_random_uuid()`
- **SSL:** disable (local) / require (produccion)

### Redis
- **Host:** localhost:6379 (local) / ElastiCache (cloud)
- **Password:** farmanexo2026 (local) / `${REDIS_PASSWORD}` (cloud)
- **DB:** 0 (compartido)
- **Uso en este servicio:**
  - No utilizado directamente actualmente
  - Planificado para: cache de perfiles, rate limiting

### LocalStack (Local) / AWS (Cloud)
- **Endpoint:** http://localhost:4566 (local)
- **Region:** us-east-1
- **Credenciales (local):** test/test (fake)

---

## RECURSOS AWS UTILIZADOS

### S3 Buckets

**Bucket:** `farmanexo-avatars`
**Uso:** Almacenamiento de avatares de usuario

**Estructura:**
```
farmanexo-avatars/
  users/
    {user_id}/
      avatar.jpg    (o .png, .webp)
```

**Operaciones:**
- **Upload:** Recibe imagen multipart, redimensiona a 512x512, comprime a 85% calidad, sube a S3
- **Download:** URL directa al objeto S3 (almacenada en `profiles.avatar_url`)
- **Delete:** Elimina objeto de S3 y limpia `avatar_url` del perfil

**URLs generadas:**
- **Local:** `http://localhost:4566/farmanexo-avatars/users/{user_id}/avatar.{ext}`
- **Cloud:** `https://farmanexo-avatars.s3.{region}.amazonaws.com/users/{user_id}/avatar.{ext}`

**Restricciones:**
- Tamano maximo: 5 MB
- Tipos permitidos: `image/jpeg`, `image/jpg`, `image/png`, `image/webp`
- Dimensiones output: 512x512 px (center-fill)

### SQS Queues

**Cola que CONSUME:**

**Cola:** `farmanexo-auth-events`
- **URL (local):** `http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/farmanexo-auth-events`
- **Configuracion:**
  - Long polling: WaitTimeSeconds = 20
  - MaxNumberOfMessages = 10
  - Procesamiento idempotente

**Eventos que procesa:**

1. **USER_REGISTERED** - Cuando Auth Service registra un usuario nuevo
   - **Accion:** Crea perfil inicial con `full_name` del evento (o email como fallback)
   - **Tambien crea:** Preferencias por defecto (language="es", theme="light", notifications_enabled=true)
   - **Idempotencia:** Verifica si el perfil ya existe antes de crear

**Cola que PUBLICA:**

**Cola:** `farmanexo-user-events`
- **URL (local):** `http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/farmanexo-user-events`

**Eventos que genera:**

1. **PROFILE_UPDATED** - Cuando se actualiza el perfil
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

2. **AVATAR_CHANGED** - Cuando se sube o elimina un avatar
```json
{
  "event_type": "AVATAR_CHANGED",
  "user_id": "uuid",
  "timestamp": "2026-02-22T12:00:00Z",
  "metadata": {
    "source": "user-service",
    "action": "uploaded"
  }
}
```

3. **ADDRESS_CREATED** - Cuando se crea una direccion
```json
{
  "event_type": "ADDRESS_CREATED",
  "user_id": "uuid",
  "timestamp": "2026-02-22T12:00:00Z",
  "metadata": {
    "source": "user-service",
    "address_id": "uuid"
  }
}
```

**Patron de publicacion:** Fire-and-forget en goroutines.

---

## ESQUEMA DE BASE DE DATOS

### Tabla: `user.profiles`

```sql
CREATE TABLE "user".profiles (
    user_id UUID PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    bio TEXT,
    avatar_url VARCHAR(500),
    date_of_birth DATE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
```

**Proposito:** Perfil principal del usuario. `user_id` corresponde al `id` de `auth.users`.

**No tiene auto-increment:** El `user_id` viene del Auth Service (UUID).

### Tabla: `user.addresses`

```sql
CREATE TABLE "user".addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    label VARCHAR(50),
    street VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100) NOT NULL DEFAULT 'Peru',
    is_default BOOLEAN NOT NULL DEFAULT false,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_addresses_user_id ON "user".addresses(user_id);
CREATE INDEX idx_addresses_is_default ON "user".addresses(user_id, is_default) WHERE is_default = true;
```

**Proposito:** Direcciones de entrega del usuario. Soporta multiples direcciones con una marcada como default.

**Coordenadas:** Latitud/longitud almacenadas como DECIMAL para geocodificacion futura.

### Tabla: `user.preferences`

```sql
CREATE TABLE "user".preferences (
    user_id UUID PRIMARY KEY,
    language VARCHAR(10) NOT NULL DEFAULT 'es',
    theme VARCHAR(20) NOT NULL DEFAULT 'light',
    notifications_enabled BOOLEAN NOT NULL DEFAULT true,
    email_notifications BOOLEAN NOT NULL DEFAULT true,
    sms_notifications BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
```

**Proposito:** Configuracion de preferencias del usuario. Se crea automaticamente al recibir evento `USER_REGISTERED`.

**Valores por defecto:** Espanol, tema claro, notificaciones habilitadas (email si, SMS no).

---

## CONFIGURACION POR AMBIENTE

### Local (config.local.yaml)
```yaml
environment: local
server:
  port: 4002
  read_timeout: 15s
  write_timeout: 15s
database:
  host: localhost
  user: admin
  password: admin
  db_name: user_db
  schema: "user"
  sslmode: disable
  max_open_conns: 25
jwt:
  secret: "dev-super-secret-key-change-in-production-min-32-chars"
aws:
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

### Development (config.development.yaml)
```yaml
environment: development
database:
  host: ${DB_HOST}
  user: ${DB_USER}
  password: ${DB_PASSWORD}
  sslmode: require
jwt:
  secret: ${JWT_SECRET}
aws:
  endpoint: ""  # AWS real
log:
  level: info
  encoding: json
```

### Production (config.production.yaml)
```yaml
environment: production
database:
  max_open_conns: 50
  max_idle_conns: 10
  sslmode: require
log:
  level: error
  encoding: json
```

---

## SECRETS Y CREDENCIALES

### Secrets Manager (Produccion)
- `farmanexo/auth/jwt-secret` - JWT validation key (mismo que Auth Service)
- `farmanexo/database/password` - Database password

### Variables de Entorno (Ambientes desplegados)

| Variable | Descripcion |
|---|---|
| `ENV` | local, development, qa, uat, production |
| `DB_HOST` | Host PostgreSQL |
| `DB_USER` | Usuario PostgreSQL |
| `DB_PASSWORD` | Password PostgreSQL |
| `JWT_SECRET` | JWT secret (mismo que Auth Service) |
| `AWS_REGION` | Region AWS |
| `SQS_USER_EVENTS_QUEUE_URL` | URL cola para publicar |
| `SQS_AUTH_EVENTS_QUEUE_URL` | URL cola para consumir |

---

## CACHE REDIS - PATRONES

User Service actualmente no utiliza Redis directamente. Patrones planificados para futuro:

| Pattern | TTL | Uso (planificado) |
|---------|-----|-----|
| `cache:user:profile:{user_id}` | 30 min | Cache de perfil |
| `cache:user:preferences:{user_id}` | 1 hora | Cache de preferencias |

---

## EVENTOS SQS

### Flujo de Eventos

```
[Auth Service] --USER_REGISTERED--> [farmanexo-auth-events] --> [User Service] (crea perfil + preferencias)
[User Service] --PROFILE_UPDATED--> [farmanexo-user-events] --> [Consumidores futuros]
[User Service] --AVATAR_CHANGED---> [farmanexo-user-events] --> [Consumidores futuros]
[User Service] --ADDRESS_CREATED--> [farmanexo-user-events] --> [Consumidores futuros]
```

### Procesamiento del Consumer
- **Tipo:** Long polling (20s wait)
- **Batch:** Hasta 10 mensajes por receive
- **Goroutine:** Corre en background desde main.go
- **Idempotencia:** Verifica existencia antes de crear
- **Error handling:** Log error y continua con siguiente mensaje

---

## DESPLIEGUE

### Checklist Pre-Deploy
- [ ] Migraciones aplicadas
- [ ] Variables de entorno configuradas
- [ ] JWT_SECRET identico al de Auth Service
- [ ] S3 bucket `farmanexo-avatars` creado
- [ ] SQS queues creadas (`farmanexo-user-events`, `farmanexo-auth-events`)
- [ ] PostgreSQL accesible con database `user_db`

### Comandos de Deploy
```bash
# Build
make build

# Migraciones
make migrate-up ENV=production

# Docker
make docker-build
make docker-run
```

---

## TESTING LOCAL

### 1. Levantar Infraestructura
```bash
cd FarmaNexo/Helpers
./start-local.sh --full
./init-localstack-resources.sh
```

### 2. Verificar Servicios
```bash
# PostgreSQL
docker exec -it farmanexo-postgres psql -U admin -d user_db

# LocalStack S3
aws --endpoint-url=http://localhost:4566 s3 ls s3://farmanexo-avatars/

# LocalStack SQS
aws --endpoint-url=http://localhost:4566 sqs list-queues
```

### 3. Crear Base de Datos
```bash
docker exec -it farmanexo-postgres psql -U admin -c "CREATE DATABASE user_db;"
```

### 4. Ejecutar Migraciones
```bash
cd services/user-service
make migrate-up
```

### 5. Ejecutar Servicio
```bash
make dev
```

---

## MONITOREO

### Metricas Importantes
- Tasa de actualizacion de perfiles
- Tamano promedio de avatares subidos
- Eventos consumidos (USER_REGISTERED) / tasa de error
- Latencia de upload a S3
- Cola SQS depth (mensajes sin procesar)

### Logs
- **Formato:** Console (local/dev), JSON (produccion)
- **Logger:** Zap structured logging
- **Campos contextuales:** user_id, correlation_id, event_type, file_size

### Alertas Recomendadas
- Rate de errores 5xx > 1%
- Cola SQS con mensajes sin procesar > 100
- S3 upload failures
- Consumer de eventos detenido

---

## TROUBLESHOOTING

### Problema: "Perfil no encontrado" despues de registrarse
**Sintoma:** GET /api/v1/users/me retorna 404 inmediatamente despues del registro
**Causa:** El consumer de USER_REGISTERED aun no proceso el evento
**Solucion:** Esperar unos segundos. Verificar que el consumer esta corriendo y la cola SQS existe.

### Problema: "Avatar upload failed"
**Sintoma:** Error al subir avatar
**Causa:** S3 bucket no existe o LocalStack no esta corriendo
**Solucion:**
```bash
aws --endpoint-url=http://localhost:4566 s3 mb s3://farmanexo-avatars
```

### Problema: "Consumer no recibe eventos"
**Sintoma:** Perfiles no se crean automaticamente al registrar usuarios
**Causa:** Cola `farmanexo-auth-events` no existe o URL incorrecta en config
**Solucion:**
```bash
aws --endpoint-url=http://localhost:4566 sqs list-queues
# Verificar que farmanexo-auth-events aparece
# Si no existe:
cd FarmaNexo/Helpers && ./init-localstack-resources.sh
```

### Problema: "Imagen demasiado grande"
**Sintoma:** Error 400 al subir avatar
**Causa:** Archivo excede 5 MB
**Solucion:** Reducir tamano de imagen antes de subir. El servicio solo acepta max 5 MB.

---

## Referencias

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [AWS S3 Documentation](https://docs.aws.amazon.com/s3/)
- [AWS SQS Documentation](https://docs.aws.amazon.com/sqs/)
- [LocalStack Documentation](https://docs.localstack.cloud/)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [GORM Documentation](https://gorm.io/docs/)
- [disintegration/imaging](https://github.com/disintegration/imaging)

---

Ultima actualizacion: 2026-02-22
