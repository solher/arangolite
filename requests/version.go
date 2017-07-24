package requests

// GetVersion returns the server version number.
type GetVersion struct {
	Details bool
}

func (r *GetVersion) Path() string {
	path := "/_api/version"

	if r.Details {
		path += "?details=true"
	}

	return path
}

func (r *GetVersion) Method() string {
	return "GET"
}

func (r *GetVersion) Generate() []byte {
	return []byte{}
}

type GetVersionResult struct {
	Server  string            `json:"server"`
	Version string            `json:"version"`
	License string            `json:"license"`
	Details map[string]string `json:"details"`
}
