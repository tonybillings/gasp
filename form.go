package gasp

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
	"strings"
	"tonysoft.com/gasp/resources"
)

const (
	defaultSocket = "127.0.0.1:8800"
)

type Form struct {
	server               *Server
	viewName             string
	html                 string
	buttonCount          int
	textboxCount         int
	labelCount           int
	dropdownCount        int
	checkboxCount        int
	linechartCount       int
	packetInspectorCount int
	ErrorChan            chan error
	Data                 map[string]interface{}
}

type FormOptions struct {
	Socket string
	Path   string
}

func NewForm(options ...FormOptions) *Form {
	if options == nil || len(options) == 0 {
		options = make([]FormOptions, 1)
	}

	socket := options[0].Socket
	path := options[0].Path

	if socket == "" {
		socket = defaultSocket
	}

	server, err := NewServer(socket)
	if err != nil {
		panic(err)
	}

	form := Form{
		html:      "",
		server:    server,
		viewName:  path,
		ErrorChan: make(chan error),
		Data:      make(map[string]interface{}),
	}
	server.form = &form

	go func() {
		for {
			serverErr := <-server.ErrorChan
			if serverErr != nil {
				form.ErrorChan <- serverErr
				if errors.Is(serverErr, http.ErrServerClosed) || websocket.IsCloseError(serverErr, websocket.CloseGoingAway) {
					return
				}
			}
		}
	}()

	form.startFormHtml()

	return &form
}

func (form *Form) GetUri() string {
	if form.server.useTls {
		return "https://" + form.server.commSocket + "/" + form.viewName
	} else {
		return "http://" + form.server.commSocket + "/" + form.viewName
	}
}

func (form *Form) PrintUri() {
	fmt.Printf("ACCESS THE UI VIA: %s", form.GetUri())
}

func (form *Form) AddResources(directory string, excludedFileExtensions ...string) *Form {
	err := form.server.AddResources(directory, excludedFileExtensions...)
	if err != nil {
		panic(err)
	}
	return form
}

func (form *Form) AddColumn(attributes ...ControlAttribute) *Form {
	atts := getAttributesHtml(attributes...)
	form.html += fmt.Sprintf("</td><td %s class=\"gtabledata\">", atts)
	return form
}

func (form *Form) AddVariableSetter(variableName string, setter func(req *http.Request) string) *Form {
	err := form.server.AddVariableSetter(variableName, setter)
	if err != nil {
		panic(err)
	}
	return form
}

func (form *Form) AddVariableLabel(variableName string, attributes ...ControlAttribute) *Form {
	id := getElementAttributeFromArray("id", attributes...)
	if id == nil {
		id = &ControlAttribute{Key: "id", Value: "glabel" + strconv.Itoa(form.labelCount)}
		attributes = append(attributes, *id)
	}
	atts := getAttributesHtml(attributes...)

	form.html += fmt.Sprintf("<label %s class=\"glabel\"><!--%s--></label><br/><br/>", atts, variableName)
	form.labelCount++
	return form
}

func (form *Form) AddTextbox(label string, attributes ...ControlAttribute) *Form {
	id := getElementAttributeFromArray("id", attributes...)
	if id == nil {
		id = &ControlAttribute{Key: "id", Value: "gtextbox" + strconv.Itoa(form.textboxCount)}
		attributes = append(attributes, *id)
	}
	atts := getAttributesHtml(attributes...)

	if label != "" {
		form.html += fmt.Sprintf("<label class=\"glabel\" for=\"%s\">%s</label>", id.Value, label)
	}

	form.html += fmt.Sprintf("<input %s type=\"text\" class=\"gtextbox\" /><br/><br/>", atts)
	form.textboxCount++

	return form
}

func (form *Form) AddButton(text string, clickHandler func(event *ClientEvent), attributes ...ControlAttribute) *Form {
	id := getElementAttributeFromArray("id", attributes...)
	if id == nil {
		id = &ControlAttribute{Key: "id", Value: "gbutton" + strconv.Itoa(form.buttonCount)}
		attributes = append(attributes, *id)
	}
	atts := getAttributesHtml(attributes...)

	form.server.AddEventHandler(form.viewName, id.Value, "click", clickHandler)

	form.html += fmt.Sprintf("<button %s class=\"gbutton\" >%s</button><br/><br/>", atts, text)
	form.buttonCount++

	return form
}

func (form *Form) AddLabel(text string, attributes ...ControlAttribute) *Form {
	id := getElementAttributeFromArray("id", attributes...)
	if id == nil {
		id = &ControlAttribute{Key: "id", Value: "glabel" + strconv.Itoa(form.labelCount)}
		attributes = append(attributes, *id)
	}
	atts := getAttributesHtml(attributes...)

	form.html += fmt.Sprintf("<label %s class=\"glabel\" >%s</label><br/><br/>", atts, text)
	form.labelCount++

	return form
}

func (form *Form) AddDropdown(label string, items []string, attributes ...ControlAttribute) *Form {
	id := getElementAttributeFromArray("id", attributes...)
	if id == nil {
		id = &ControlAttribute{Key: "id", Value: "gdropdown" + strconv.Itoa(form.dropdownCount)}
		attributes = append(attributes, *id)
	}
	atts := getAttributesHtml(attributes...)

	if label != "" {
		form.html += fmt.Sprintf("<label class=\"glabel\" for=\"%s\">%s</label>", id.Value, label)
	}

	itemsHtml := ""
	for _, item := range items {
		itemsHtml += fmt.Sprintf("<option value=\"%s\">%s</option>", item, item)
	}

	form.html += fmt.Sprintf("<select %s class=\"gdropdown\" >%s</select><br/><br/>", atts, itemsHtml)
	form.dropdownCount++

	return form
}

func (form *Form) AddCheckbox(label string, attributes ...ControlAttribute) *Form {
	id := getElementAttributeFromArray("id", attributes...)
	if id == nil {
		id = &ControlAttribute{Key: "id", Value: "gcheckbox" + strconv.Itoa(form.checkboxCount)}
		attributes = append(attributes, *id)
	}
	atts := getAttributesHtml(attributes...)

	lineBreak := ""
	if label == "" {
		lineBreak = "<br/><br/>"
	}
	form.html += fmt.Sprintf("<input %s type=\"checkbox\" class=\"gcheckbox\"/>%s", atts, lineBreak)

	if label != "" {
		form.html += fmt.Sprintf("<label class=\"glabel\" for=\"%s\">%s</label><br/><br/>", id.Value, label)
	}

	form.checkboxCount++
	return form
}

func (form *Form) AddLineChart(initialState LineChartState, attributes ...ControlAttribute) *Form {
	var idAtt *ControlAttribute
	if initialState.Id == "" {
		idAtt = getElementAttributeFromArray("id", attributes...)
		if idAtt == nil {
			idAtt = &ControlAttribute{Key: "id", Value: "glinechart" + strconv.Itoa(form.linechartCount)}
		}
	} else {
		idAtt = &ControlAttribute{Key: "id", Value: initialState.Id}
	}
	attributes = append(attributes, *idAtt)

	if initialState.Width == 0 {
		initialState.Width = 500
	}
	if initialState.Height == 0 {
		initialState.Height = 200
	}

	if initialState.Lines == nil || len(initialState.Lines) == 0 {
		initialState.Lines = []*LineState{{}}
	}

	atts := getAttributesHtml(attributes...)
	initialStateJson, err := json.Marshal(initialState)
	if err != nil {
		panic(err)
	}
	initialStateEncoded := base64.StdEncoding.EncodeToString(initialStateJson)
	form.html += fmt.Sprintf("<canvas %s class=\"glinechart\" width=\"%d\" height=\"%d\" data-initial-state=\"%s\"></canvas>", atts, initialState.Width, initialState.Height, initialStateEncoded)
	form.linechartCount++
	return form
}

func (form *Form) AddPacketInspector(initialState PacketInspectorState, attributes ...ControlAttribute) *Form {
	var idAtt *ControlAttribute
	if initialState.Id == "" {
		idAtt = getElementAttributeFromArray("id", attributes...)
		if idAtt == nil {
			idAtt = &ControlAttribute{Key: "id", Value: "gpacketinspector" + strconv.Itoa(form.packetInspectorCount)}
		}
	} else {
		idAtt = &ControlAttribute{Key: "id", Value: initialState.Id}
	}
	attributes = append(attributes, *idAtt)

	if initialState.Width == 0 {
		initialState.Width = 500
	}

	atts := getAttributesHtml(attributes...)
	initialStateJson, err := json.Marshal(initialState)
	if err != nil {
		panic(err)
	}

	initialStateEncoded := base64.StdEncoding.EncodeToString(initialStateJson)
	form.html += fmt.Sprintf("<div %s style=\"width:%dpx;\" class=\"gpacketinspector\" data-initial-state=\"%s\"><canvas width=\"%d\" height=\"275\"></canvas></div>", atts, initialState.Width, initialStateEncoded, initialState.Width)
	form.packetInspectorCount++
	return form
}

func (form *Form) AddEventHandler(view string, elementId string, eventType string, handler func(event *ClientEvent)) *Form {
	form.server.AddEventHandler(view, elementId, eventType, handler)
	return form
}

func (form *Form) GetTextbox(id ...string) *Textbox {
	control := Textbox{}

	idVal := "gtextbox0"
	if id != nil && len(id) > 0 {
		idVal = id[0]
	}
	control.id = idVal

	control.form = form
	return &control
}

func (form *Form) GetButton(id ...string) *Button {
	control := Button{}

	idVal := "gbutton0"
	if id != nil && len(id) > 0 && id[0] != "" {
		idVal = id[0]
	}
	control.id = idVal

	control.form = form
	return &control
}

func (form *Form) GetLabel(id ...string) *Label {
	control := Label{}

	idVal := "glabel0"
	if id != nil && len(id) > 0 && id[0] != "" {
		idVal = id[0]
	}
	control.id = idVal

	control.form = form
	return &control
}

func (form *Form) GetDropdown(id ...string) *Dropdown {
	control := Dropdown{}

	idVal := "gdropdown0"
	if id != nil && len(id) > 0 && id[0] != "" {
		idVal = id[0]
	}
	control.id = idVal

	control.form = form
	return &control
}

func (form *Form) GetCheckbox(id ...string) *Checkbox {
	control := Checkbox{}

	idVal := "gcheckbox0"
	if id != nil && len(id) > 0 && id[0] != "" {
		idVal = id[0]
	}
	control.id = idVal

	control.form = form
	return &control
}

func (form *Form) GetLineChart(id ...string) *LineChart {
	control := LineChart{}

	idVal := "glinechart0"
	if id != nil && len(id) > 0 && id[0] != "" {
		idVal = id[0]
	}
	control.id = idVal

	control.form = form
	return &control
}

func (form *Form) GetPacketInspector(id ...string) *PacketInspector {
	control := PacketInspector{}

	idVal := "gpacketinspector0"
	if id != nil && len(id) > 0 && id[0] != "" {
		idVal = id[0]
	}
	control.id = idVal

	control.form = form
	return &control
}

func (form *Form) Start() (*Form, error) {
	form.endFormHtml()
	html := strings.ReplaceAll(resources.FormTemplate, "<!--form-->", form.html)
	err := form.server.AddView(form.viewName, html)
	if err != nil {
		return form, err
	}
	return form, form.server.Start()
}

func (form *Form) StartWithTLS(certFile string, keyFile string) error {
	form.endFormHtml()
	html := strings.ReplaceAll(resources.FormTemplate, "<!--form-->", form.html)
	err := form.server.AddView(form.viewName, html)
	if err != nil {
		return err
	}
	return form.server.StartWithTLS(certFile, keyFile)
}

func (form *Form) Stop() error {
	return form.server.Stop()
}

func (form *Form) Update(state *FormState) {
	evt := ServerEvent{
		Type: "form_update",
		Text: "the form has been updated server-side",
		Data: map[string]interface{}{"state": state},
	}

	form.server.SendEvent(&evt)
}

func (form *Form) UpdateTextbox(state *TextboxState, propertiesToUpdate ...string) {
	evt := ServerEvent{
		Type: "textbox_update",
		Text: "the textbox has been updated server-side",
		Data: map[string]interface{}{
			"state":      state,
			"properties": propertiesToUpdate,
		},
	}

	form.server.SendEvent(&evt)
}

func (form *Form) UpdateButton(state *ButtonState, propertiesToUpdate ...string) {
	evt := ServerEvent{
		Type: "button_update",
		Text: "the button has been updated server-side",
		Data: map[string]interface{}{
			"state":      state,
			"properties": propertiesToUpdate,
		},
	}

	form.server.SendEvent(&evt)
}

func (form *Form) UpdateLabel(state *LabelState, propertiesToUpdate ...string) {
	evt := ServerEvent{
		Type: "label_update",
		Text: "the label has been updated server-side",
		Data: map[string]interface{}{
			"state":      state,
			"properties": propertiesToUpdate,
		},
	}

	form.server.SendEvent(&evt)
}

func (form *Form) UpdateDropdown(state *DropdownState, propertiesToUpdate ...string) {
	evt := ServerEvent{
		Type: "dropdown_update",
		Text: "the drop-down has been updated server-side",
		Data: map[string]interface{}{
			"state":      state,
			"properties": propertiesToUpdate,
		},
	}

	form.server.SendEvent(&evt)
}

func (form *Form) UpdateCheckbox(state *CheckboxState, propertiesToUpdate ...string) {
	evt := ServerEvent{
		Type: "checkbox_update",
		Text: "the checkbox has been updated server-side",
		Data: map[string]interface{}{
			"state":      state,
			"properties": propertiesToUpdate,
		},
	}

	form.server.SendEvent(&evt)
}

func (form *Form) UpdateLineChart(state *LineChartState, propertiesToUpdate ...string) {
	evt := ServerEvent{
		Type: "linechart_update",
		Text: "the linechart has been updated server-side",
		Data: map[string]interface{}{
			"state":      state,
			"properties": propertiesToUpdate,
		},
	}

	form.server.SendEvent(&evt)
}

func (form *Form) UpdatePacketInspector(state *PacketInspectorState, propertiesToUpdate ...string) {
	evt := ServerEvent{
		Type: "packetinspector_update",
		Text: "the packetinspector has been updated server-side",
		Data: map[string]interface{}{
			"state":      state,
			"properties": propertiesToUpdate,
		},
	}

	form.server.SendEvent(&evt)
}

func (form *Form) startFormHtml() {
	form.html = "<div class=\"gform\"><table class=\"gtable\"><tr><td class=\"gtabledata\">"
}

func (form *Form) endFormHtml() {
	form.html += "</td></tr></div>"
}

func getAttributesHtml(attributes ...ControlAttribute) string {
	html := ""
	for _, att := range attributes {
		html += att.Key + "=\"" + att.Value + "\" "
	}
	return html
}

func getElementAttributeFromArray(key string, attributes ...ControlAttribute) *ControlAttribute {
	for _, attr := range attributes {
		if attr.Key == key {
			return &attr
		}
	}
	return nil
}
