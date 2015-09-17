package arangolite

type Model struct {
	SysID  string `json:"_id,omitempty"`
	SysRev string `json:"_rev,omitempty"`
	SysKey string `json:"_key,omitempty"`
	ID     string `json:"id,omitempty"`
}

func (m *Model) RewriteIDs() {
	m.ID = m.SysKey
	m.SysID = ""
	m.SysRev = ""
	m.SysKey = ""
}
