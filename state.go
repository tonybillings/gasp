package gasp

type ControlAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ControlState struct {
	Id        string `json:"id"`
	Text      string `json:"text"`
	IsVisible bool   `json:"is_visible"`
	IsEnabled bool   `json:"is_enabled"`
}

type TextboxState struct {
	ControlState
}

type ButtonState struct {
	ControlState
}

type LabelState struct {
	ControlState
}

type DropdownState struct {
	ControlState
}

type CheckboxState struct {
	ControlState
	IsChecked bool `json:"is_checked"`
}

type LineState struct {
	Name      string    `json:"name"`
	Thickness int       `json:"thickness"`
	Color     string    `json:"color"`
	NewValues []float64 `json:"new_values"`
}

type LineChartState struct {
	ControlState
	Width  int          `json:"width"`
	Height int          `json:"height"`
	Lines  []*LineState `json:"lines"`
}

type PacketInspectorState struct {
	ControlState
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	CharHeight int    `json:"char_height"`
	ByteCount  int    `json:"byte_count"`
	Packet     string `json:"packet"` // base64-encoded packet bytes
}

type FormState struct {
	Textboxes        []*TextboxState         `json:"textboxes"`
	Buttons          []*ButtonState          `json:"buttons"`
	Labels           []*LabelState           `json:"labels"`
	Dropdowns        []*DropdownState        `json:"dropdowns"`
	Checkboxes       []*CheckboxState        `json:"checkboxes"`
	LineCharts       []*LineChartState       `json:"linecharts"`
	PacketInspectors []*PacketInspectorState `json:"packetinspectors"`
}

func (formState *FormState) GetTextbox(id string) *TextboxState {
	for _, s := range formState.Textboxes {
		if s.Id == id {
			return s
		}
	}
	return nil
}

func (formState *FormState) GetButton(id string) *ButtonState {
	for _, s := range formState.Buttons {
		if s.Id == id {
			return s
		}
	}
	return nil
}

func (formState *FormState) GetLabel(id string) *LabelState {
	for _, s := range formState.Labels {
		if s.Id == id {
			return s
		}
	}
	return nil
}

func (formState *FormState) GetDropdown(id string) *DropdownState {
	for _, s := range formState.Dropdowns {
		if s.Id == id {
			return s
		}
	}
	return nil
}

func (formState *FormState) GetCheckbox(id string) *CheckboxState {
	for _, s := range formState.Checkboxes {
		if s.Id == id {
			return s
		}
	}
	return nil
}

func (formState *FormState) GetLineChart(id string) *LineChartState {
	for _, s := range formState.LineCharts {
		if s.Id == id {
			return s
		}
	}
	return nil
}

func (formState *FormState) GetPacketInspector(id string) *PacketInspectorState {
	for _, s := range formState.PacketInspectors {
		if s.Id == id {
			return s
		}
	}
	return nil
}
