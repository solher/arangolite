package arangolite

type Model struct {
	AraID  string `json:"_id,omitempty"`
	AraRev string `json:"_rev,omitempty"`
	AraKey string `json:"_key,omitempty"`
	ID     string `json:"id,omitempty"`
}

func (m *Model) RewriteIDs() {
	m.ID = m.AraID
	m.AraID = ""
	m.AraRev = ""
	m.AraKey = ""
}
