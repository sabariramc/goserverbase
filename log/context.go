package log

type ContextVariable string

const (
	ContextKeyCorrelation        ContextVariable = "correlationParam"
	ContextKeyCustomerIdentifier ContextVariable = "customerIdentity"
)
