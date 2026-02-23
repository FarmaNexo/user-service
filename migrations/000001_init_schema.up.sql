-- migrations/000001_init_schema.up.sql
-- Migración inicial del User Service

-- Crear esquema dedicado para user
CREATE SCHEMA IF NOT EXISTS "user";

-- Habilitar extensión UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Tabla profiles en esquema user
CREATE TABLE IF NOT EXISTS "user".profiles (
    user_id UUID PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    bio TEXT,
    avatar_url VARCHAR(500),
    date_of_birth DATE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Tabla addresses en esquema user
CREATE TABLE IF NOT EXISTS "user".addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    label VARCHAR(50),
    street VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100) NOT NULL DEFAULT 'Perú',
    is_default BOOLEAN NOT NULL DEFAULT false,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Índices para addresses
CREATE INDEX idx_addresses_user_id ON "user".addresses(user_id);
CREATE INDEX idx_addresses_is_default ON "user".addresses(user_id, is_default) WHERE is_default = true;

-- Tabla preferences en esquema user
CREATE TABLE IF NOT EXISTS "user".preferences (
    user_id UUID PRIMARY KEY,
    language VARCHAR(10) NOT NULL DEFAULT 'es',
    theme VARCHAR(20) NOT NULL DEFAULT 'light',
    notifications_enabled BOOLEAN NOT NULL DEFAULT true,
    email_notifications BOOLEAN NOT NULL DEFAULT true,
    sms_notifications BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Comentarios en tablas
COMMENT ON TABLE "user".profiles IS 'Perfiles de usuario del sistema';
COMMENT ON TABLE "user".addresses IS 'Direcciones de los usuarios';
COMMENT ON TABLE "user".preferences IS 'Preferencias de los usuarios';

-- Comentarios en columnas importantes
COMMENT ON COLUMN "user".profiles.user_id IS 'Referencia al ID del usuario en auth.users (NO FK, servicio separado)';
COMMENT ON COLUMN "user".profiles.avatar_url IS 'URL del avatar en S3';
COMMENT ON COLUMN "user".addresses.is_default IS 'Solo una dirección puede ser default por usuario';
COMMENT ON COLUMN "user".preferences.language IS 'Idioma: es, en';
COMMENT ON COLUMN "user".preferences.theme IS 'Tema: light, dark';
