package facebook

type ResponseData struct {
	ApplID           string            `json:"app_id,omitempty"`
	Application      string            `json:"application,omitempty"`
	ExpiresAfterSecs int               `json:"expires_at,omitempty"`
	Valid            bool              `json:"is_valid,omitempty"`
	IssuedBeforeSecs int               `json:"issued_at,omitempty"`
	Meta             map[string]string `json:"metadata,omitempty"`
	FBScopes         []string          `json:"scopes,omitempty"`
	UsrID            string            `json:"user_id,omitempty"`
}

type Response struct {
	ResponseData `json:"data,omitempty"`
}

func (t *Response) UserID() string {
	if t == nil {
		return ""
	}
	return t.UsrID
}

func (t *Response) Issued() int {
	if t == nil {
		return -1
	}
	return t.IssuedBeforeSecs
}

func (t *Response) Expiry() int {
	if t == nil {
		return -1
	}
	return t.ExpiresAfterSecs
}

func (t *Response) AppID() string {
	if t == nil {
		return ""
	}
	return t.ApplID
}

func (t *Response) AppName() string {
	if t == nil {
		return ""
	}
	return t.Application
}

func (t *Response) Scopes() []string {
	if t == nil {
		return nil
	}
	return t.FBScopes
}

func (t *Response) Metadata() map[string]string {
	if t == nil {
		return nil
	}
	return t.Meta
}
