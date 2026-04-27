CREATE TABLE xui_access (
    id BIGSERIAL PRIMARY KEY,

    client_uuid UUID NOT NULL,
    inbound_id BIGINT NOT NULL,

    xui_client_email TEXT NOT NULL,
    vless_uri TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_xui_access_client_uuid
        FOREIGN KEY (client_uuid) REFERENCES vpn_clients (client_uuid)
        ON DELETE CASCADE,

    CONSTRAINT uq_xui_access_client_uuid UNIQUE (client_uuid)
);

CREATE INDEX idx_xui_access_inbound_id ON xui_access (inbound_id);