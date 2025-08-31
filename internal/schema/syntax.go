package schema

type Syntax struct {
	Platform string
	Method   string
	Endpoint string // Method without custom request parameters
	Id       string
}
