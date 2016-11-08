package user

type value struct {
	value     string
	validated bool
}

func (v value) Value() string {
	return v.value
}

func (v value) Validated() bool {
	return v.validated
}
