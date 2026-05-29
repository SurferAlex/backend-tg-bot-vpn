ALTER TABLE vpn_clients DROP CONSTRAINT IF EXISTS fk_vpn_clients_server;
DROP INDEX IF EXISTS idx_vpn_clients_server_id;
ALTER TABLE vpn_clients DROP COLUMN IF EXISTS server_id;
DROP TABLE IF EXISTS vpn_servers;
