package gasp

type ServerEvent struct {
	Type string                 `json:"type"`
	Text string                 `json:"text"`
	Data map[string]interface{} `json:"data,omitempty"`
}

type ClientEvent struct {
	View  string                 `json:"view"`
	Id    string                 `json:"id"`
	Type  string                 `json:"type"`
	Data  map[string]interface{} `json:"data,omitempty"`
	State FormState              `json:"state"`
	Form  *Form                  `json:"-"`
}
