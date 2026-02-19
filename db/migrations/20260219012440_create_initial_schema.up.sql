-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =====================================================
-- USERS
-- =====================================================

CREATE TABLE users (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fullname       VARCHAR(100),
    email          VARCHAR(150),
    phone          VARCHAR(30),
    birth_date     DATE,
    address        TEXT,
    subdistrict_id INT,
    district_id    INT,
    city_id        INT,
    profile_image  TEXT,
    is_active      BOOLEAN NOT NULL DEFAULT TRUE,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_by     VARCHAR(100),
    updated_by     VARCHAR(100),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);

CREATE UNIQUE INDEX ux_users_email ON users (email) WHERE email IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_users_deleted_at ON users (deleted_at) WHERE deleted_at IS NULL;

-- =====================================================
-- AUTH PROVIDERS (email | google | apple)
-- =====================================================

CREATE TABLE user_auth_providers (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL,
    provider         VARCHAR(30) NOT NULL,    -- email | google | apple
    provider_user_id TEXT,                    -- sub dari Google / Apple
    email            VARCHAR(150),
    password_hash    TEXT,                    -- hanya provider=email, simpan bcrypt hash
    created_by       VARCHAR(100),
    updated_by       VARCHAR(100),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_auth_providers_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT uq_auth_providers_provider_user UNIQUE (provider, provider_user_id)
);

CREATE INDEX idx_auth_providers_user_id ON user_auth_providers (user_id);

-- =====================================================
-- REFRESH TOKENS (JWT)
-- =====================================================

CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL,
    token_hash  TEXT NOT NULL,        -- sha256 dari refresh token plain
    device_info TEXT,                 -- user agent / device identifier
    ip_address  VARCHAR(45),
    is_revoked  BOOLEAN NOT NULL DEFAULT FALSE,
    expired_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_refresh_tokens_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens (token_hash);

-- =====================================================
-- OAUTH STATES (CSRF protection untuk Google & Apple)
-- =====================================================

CREATE TABLE oauth_states (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    state        TEXT NOT NULL UNIQUE,  -- random string dikirim ke provider
    provider     VARCHAR(30),           -- google | apple
    redirect_uri TEXT,
    expired_at   TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- =====================================================
-- OTP (email verification, login, reset password)
-- =====================================================

CREATE TABLE otps (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL,
    destination   TEXT,                 -- email / nomor HP
    otp_code      TEXT,                 -- simpan HASH (sha256)
    purpose       VARCHAR(30),          -- email_verification | login | reset_password
    expired_at    TIMESTAMPTZ,
    verified_at   TIMESTAMPTZ,
    attempt_count INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_otps_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX idx_otps_user_id ON otps (user_id);
CREATE INDEX idx_otps_destination ON otps (destination);

CREATE TABLE otp_deliveries (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    otp_id        UUID NOT NULL,
    channel       VARCHAR(20),          -- email | sms | whatsapp
    provider      VARCHAR(50),
    status        VARCHAR(20),          -- pending | sent | delivered | failed
    sent_at       TIMESTAMPTZ,
    delivered_at  TIMESTAMPTZ,
    error_message TEXT,
    CONSTRAINT fk_otp_deliveries_otp FOREIGN KEY (otp_id) REFERENCES otps (id) ON DELETE CASCADE
);

CREATE INDEX idx_otp_deliveries_otp_id ON otp_deliveries (otp_id);

-- =====================================================
-- ROLES & USER ROLES
-- =====================================================

CREATE TABLE roles (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(30) NOT NULL UNIQUE,  -- ADMIN | FRONT_DESK | HOUSEKEEPING | MANAGER | CUSTOMER
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO roles (code, name) VALUES
    ('ADMIN',        'Administrator'),
    ('MANAGER',      'Manager'),
    ('FRONT_DESK',   'Front Desk'),
    ('HOUSEKEEPING', 'Housekeeping'),
    ('CUSTOMER',     'Customer');

CREATE TABLE user_roles (
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,
    PRIMARY KEY (user_id, role_id),
    CONSTRAINT fk_user_roles_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_user_roles_role FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE
);

CREATE INDEX idx_user_roles_user_id ON user_roles (user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles (role_id);
