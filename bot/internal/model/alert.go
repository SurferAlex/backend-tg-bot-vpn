package model

type AlertRequest struct {
	ClientUUID string `json:"client_uuid"`
	IPCount    int    `json:"ip_count"`
	MaxIPs     int    `json:"max_ips"`
	Text       string `json:"text"`
}
