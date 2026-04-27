CREATE TABLE vpn_clients (
    id BIGSERIAL PRIMARY KEY,
    client_uuid UUID NOT NULL,
    telegram_user_id BIGINT,
    max_ips INT NOT NULL DEFAULT 2 CHECK (max_ips > 0),
    key_expires_at TIMESTAMPTZ NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_vpn_clients_client_uuid UNIQUE (client_uuid)
);
CREATE INDEX idx_vpn_clients_telegram_user_id ON vpn_clients (telegram_user_id)
    WHERE telegram_user_id IS NOT NULL;
    
CREATE INDEX idx_vpn_clients_key_expires_at ON vpn_clients (key_expires_at)
    WHERE is_active = true;