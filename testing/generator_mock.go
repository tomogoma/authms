package testing

type GeneratorMock struct {
	ExpSRBsErr error
	ExpSRBs    []byte
}

func (kg *GeneratorMock) SecureRandomBytes(length int) ([]byte, error) {
	return kg.ExpSRBs, kg.ExpSRBsErr
}
