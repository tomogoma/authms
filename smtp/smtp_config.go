package smtp

type Config struct {
	ServerAddress string `json:"serverAddress,omitempty"`
	TLSPort       int32  `json:"TLSPort,omitempty"`
	SSLPort       int32  `json:"SSLPort,omitempty"`
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	FromEmail     string `json:"fromEmail,omitempty"`
	CreatedAt     string `json:"createdAt,omitempty"`
	UpdatedAt     string `json:"updatedAt,omitempty"`
}
