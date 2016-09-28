package requests

// ImportCollection imports data to a collection
type ImportCollection struct {
	CollectionName string
	Data           []byte
	FromPrefix     string
	ToPrefix       string
	Overwrite      bool
	WaitForSync    bool
	OnDuplicate    string
	Complete       bool
	Details        bool
}

func (c *ImportCollection) Path() string {
	path := "/_api/import/?type=auto&collection=" + c.CollectionName

	if c.FromPrefix != "" {
		path += "&fromPrefix=" + c.FromPrefix
	}

	if c.ToPrefix != "" {
		path += "&toPrefix=" + c.ToPrefix
	}

	if c.Overwrite {
		path += "&overwrite=yes"
	}

	if c.WaitForSync {
		path += "&waitForSync=yes"
	}

	if c.OnDuplicate != "" {
		path += "&onDuplicate=" + c.OnDuplicate
	}

	if c.Complete {
		path += "&complete=yes"
	}

	if c.Details {
		path += "&details=yes"
	}

	return path
}

func (c *ImportCollection) Method() string {
	return "POST"
}

func (c *ImportCollection) Generate() []byte {
	return c.Data
}
