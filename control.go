package gasp

import "encoding/base64"

type FormControl struct {
	id   string
	form *Form
}

type Textbox struct {
	FormControl
}

type Button struct {
	FormControl
}

type Label struct {
	FormControl
}

type Dropdown struct {
	FormControl
}

type Checkbox struct {
	FormControl
}

type LineChart struct {
	FormControl
}

type PacketInspector struct {
	FormControl
}

func (control *Textbox) UpdateText(text string) {
	state := TextboxState{}
	state.Id = control.id
	state.Text = text
	control.form.UpdateTextbox(&state, "text")
}

func (control *Textbox) UpdateIsVisible(isVisible bool) {
	state := TextboxState{}
	state.Id = control.id
	state.IsVisible = isVisible
	control.form.UpdateTextbox(&state, "is_visible")
}

func (control *Textbox) UpdateIsEnabled(isEnabled bool) {
	state := TextboxState{}
	state.Id = control.id
	state.IsEnabled = isEnabled
	control.form.UpdateTextbox(&state, "is_enabled")
}

func (control *Button) UpdateText(text string) {
	state := ButtonState{}
	state.Id = control.id
	state.Text = text
	control.form.UpdateButton(&state, "text")
}

func (control *Button) UpdateIsVisible(isVisible bool) {
	state := ButtonState{}
	state.Id = control.id
	state.IsVisible = isVisible
	control.form.UpdateButton(&state, "is_visible")
}

func (control *Button) UpdateIsEnabled(isEnabled bool) {
	state := ButtonState{}
	state.Id = control.id
	state.IsEnabled = isEnabled
	control.form.UpdateButton(&state, "is_enabled")
}

func (control *Label) UpdateText(text string) {
	state := LabelState{}
	state.Id = control.id
	state.Text = text
	control.form.UpdateLabel(&state, "text")
}

func (control *Label) UpdateIsVisible(isVisible bool) {
	state := LabelState{}
	state.Id = control.id
	state.IsVisible = isVisible
	control.form.UpdateLabel(&state, "is_visible")
}

func (control *Label) UpdateIsEnabled(isEnabled bool) {
	state := LabelState{}
	state.Id = control.id
	state.IsEnabled = isEnabled
	control.form.UpdateLabel(&state, "is_enabled")
}

func (control *Dropdown) UpdateText(text string) {
	state := DropdownState{}
	state.Id = control.id
	state.Text = text
	control.form.UpdateDropdown(&state, "text")
}

func (control *Dropdown) UpdateIsVisible(isVisible bool) {
	state := DropdownState{}
	state.Id = control.id
	state.IsVisible = isVisible
	control.form.UpdateDropdown(&state, "is_visible")
}

func (control *Dropdown) UpdateIsEnabled(isEnabled bool) {
	state := DropdownState{}
	state.Id = control.id
	state.IsEnabled = isEnabled
	control.form.UpdateDropdown(&state, "is_enabled")
}

func (control *Checkbox) UpdateText(text string) {
	state := CheckboxState{}
	state.Id = control.id
	state.Text = text
	control.form.UpdateCheckbox(&state, "text")
}

func (control *Checkbox) UpdateIsVisible(isVisible bool) {
	state := CheckboxState{}
	state.Id = control.id
	state.IsVisible = isVisible
	control.form.UpdateCheckbox(&state, "is_visible")
}

func (control *Checkbox) UpdateIsEnabled(isEnabled bool) {
	state := CheckboxState{}
	state.Id = control.id
	state.IsEnabled = isEnabled
	control.form.UpdateCheckbox(&state, "is_enabled")
}

func (control *Checkbox) UpdateIsChecked(isChecked bool) {
	state := CheckboxState{}
	state.Id = control.id
	state.IsChecked = isChecked
	control.form.UpdateCheckbox(&state, "is_checked")
}

func (control *LineChart) UpdateValues(lineName string, values ...float64) {
	if lineName == "" {
		lineName = "line0"
	}

	lineState := LineState{
		Name:      lineName,
		NewValues: values,
	}
	state := LineChartState{}
	state.Id = control.id
	state.Lines = []*LineState{&lineState}
	control.form.UpdateLineChart(&state, "values")
}

func (control *LineChart) UpdateIsVisible(isVisible bool) {
	state := LineChartState{}
	state.Id = control.id
	state.IsVisible = isVisible
	control.form.UpdateLineChart(&state, "is_visible")
}

func (control *LineChart) UpdateIsEnabled(isEnabled bool) {
	state := LineChartState{}
	state.Id = control.id
	state.IsEnabled = isEnabled
	control.form.UpdateLineChart(&state, "is_enabled")
}

func (control *PacketInspector) UpdateBytes(packet []byte) {
	state := PacketInspectorState{}
	state.Id = control.id
	state.Packet = base64.StdEncoding.EncodeToString(packet)
	control.form.UpdatePacketInspector(&state, "packet")
}

func (control *PacketInspector) UpdateIsVisible(isVisible bool) {
	state := PacketInspectorState{}
	state.Id = control.id
	state.IsVisible = isVisible
	control.form.UpdatePacketInspector(&state, "is_visible")
}

func (control *PacketInspector) UpdateIsEnabled(isEnabled bool) {
	state := PacketInspectorState{}
	state.Id = control.id
	state.IsEnabled = isEnabled
	control.form.UpdatePacketInspector(&state, "is_enabled")
}
