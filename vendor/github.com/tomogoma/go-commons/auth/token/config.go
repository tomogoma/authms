package token

type ConfigStub struct {
	TknKyFile string `yaml:"tokenKeyFile,omitempty",json:"tokenKeyFile,omitempty"`
}

func (c ConfigStub) KeyFile() string {
	return c.TknKyFile
}
