-- Second VPN node (VPS #2). Run after migration 000003 if you do NOT use env bootstrap.
-- Alternative: set VPN_SERVER_VPS1_* in backend/VpnAPI/.env (see .env.example).
-- VpnAPI must reach xui_base_url over the network (control-plane variant A).

INSERT INTO vpn_servers (
    id,
    name,
    is_active,
    xui_base_url,
    xui_username,
    xui_password,
    xui_inbound_id,
    xui_external_host,
    xui_fingerprint,
    xui_spiderx,
    xui_flow,
    xui_host_header,
    xui_server_name,
    xui_insecure_skip_verify
) VALUES (
    'vps2',
    'VPS 2',
    true,
    'http://SECOND_VPS_IP:2053',
    'admin',
    'CHANGE_ME',
    1,
    'vpn.example.com',
    'chrome',
    '/',
    '',
    '',
    '',
    false
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    is_active = EXCLUDED.is_active,
    xui_base_url = EXCLUDED.xui_base_url,
    xui_username = EXCLUDED.xui_username,
    xui_password = EXCLUDED.xui_password,
    xui_inbound_id = EXCLUDED.xui_inbound_id,
    xui_external_host = EXCLUDED.xui_external_host,
    updated_at = now();
