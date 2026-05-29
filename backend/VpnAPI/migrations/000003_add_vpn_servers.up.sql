CREATE TABLE vpn_servers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    xui_base_url TEXT NOT NULL,
    xui_username TEXT NOT NULL,
    xui_password TEXT NOT NULL,
    xui_inbound_id BIGINT NOT NULL,
    xui_external_host TEXT NOT NULL,
    xui_fingerprint TEXT NOT NULL DEFAULT 'chrome',
    xui_spiderx TEXT NOT NULL DEFAULT '/',
    xui_flow TEXT NOT NULL DEFAULT '',
    xui_host_header TEXT NOT NULL DEFAULT '',
    xui_server_name TEXT NOT NULL DEFAULT '',
    xui_insecure_skip_verify BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Placeholder row; VpnAPI overwrites from XUI_* env on startup.
INSERT INTO vpn_servers (
    id, name, xui_base_url, xui_username, xui_password, xui_inbound_id, xui_external_host
) VALUES (
    'default', 'Default', 'http://127.0.0.1', 'admin', 'admin', 1, '127.0.0.1'
);

ALTER TABLE vpn_clients ADD COLUMN server_id TEXT;

UPDATE vpn_clients SET server_id = 'default' WHERE server_id IS NULL;

ALTER TABLE vpn_clients ALTER COLUMN server_id SET NOT NULL;

ALTER TABLE vpn_clients
    ADD CONSTRAINT fk_vpn_clients_server
    FOREIGN KEY (server_id) REFERENCES vpn_servers (id);

CREATE INDEX idx_vpn_clients_server_id ON vpn_clients (server_id);
