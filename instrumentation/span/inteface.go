package span

type Span interface {
	SetTag(name string, value string)
	SetError(err error)
	Finish()
}
