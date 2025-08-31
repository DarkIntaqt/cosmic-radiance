package schema

type InvalidPathSyntaxError struct {
	path string
}

func (e *InvalidPathSyntaxError) Error() string {
	return "Invalid path syntax " + e.path
}
