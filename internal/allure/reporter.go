package allure

type Reporter struct {
	Config MaskingConfig
}

func NewReporter(config MaskingConfig) *Reporter {
	return &Reporter{Config: config}
}

func NewDefaultReporter() *Reporter {
	return NewReporter(DefaultConfig())
}
