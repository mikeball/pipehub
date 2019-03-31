package internal

// Pipe holds the pipe configuration.
type Pipe struct {
	ImportPath      string
	ImportPathAlias string
	Module          string
	Version         string
	Config          map[string]interface{}
}

type Host struct {
	Endpoint string
	Handler  string
}
