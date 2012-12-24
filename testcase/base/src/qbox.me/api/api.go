package api

type Env struct {
	Id string `json:id`

	Hosts map[string]string `json:hosts`
	Ips   map[string]string `json:"ips`

	Fopd      string `json:"fopd"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}
