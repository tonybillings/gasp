package gasp_test

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
	ui "tonysoft.com/gasp"
)

var (
	socket = "127.0.0.1:8800"
)

// NOTE: THESE TESTS MUST BE RUN INDIVIDUALLY SINCE
// THE http PACKAGE DOES NOT PROVIDE A WAY TO REMOVE
// A HANDLER AFTER CALLING http.HandleFunc()!

func TestServerStartStop(t *testing.T) {
	server := getNewServer(t)
	if server == nil {
		return
	}

	handleErrorChannel(t, server.ErrorChan)

	err := server.Start()
	if err != nil {
		t.Error(err)
		return
	}

	resp, err := getResponse("http://" + socket)
	if err != nil {
		t.Error(err)
		return
	}

	if resp != "Gasp Server Online" {
		t.Errorf("invalid response received: %s", resp)
		return
	}

	err = server.Stop()
	if err != nil {
		t.Error(err)
		return
	}

	err = server.Start()
	if err != nil {
		t.Error(err)
		return
	}

	err = server.Stop()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestCustomHandler(t *testing.T) {
	server := getNewServer(t)
	if server == nil {
		return
	}

	handleErrorChannel(t, server.ErrorChan)

	err := server.AddRouteHandler("test", func(rw http.ResponseWriter, req *http.Request) {
		_, err := rw.Write([]byte("abc123"))
		if err != nil {
			t.Error(err)
			return
		}
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = server.Start()
	if err != nil {
		t.Error(err)
		return
	}

	resp, err := getResponse("http://" + socket + "/test")
	if err != nil {
		t.Error(err)
		return
	}

	if resp != "abc123" {
		t.Errorf("invalid response received: %s", resp)
		return
	}

	err = server.Stop()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestCustomView(t *testing.T) {
	server := getNewServer(t)
	if server == nil {
		return
	}

	handleErrorChannel(t, server.ErrorChan)

	err := server.AddView("test", "xyz456")
	if err != nil {
		t.Error(err)
		return
	}

	err = server.Start()
	if err != nil {
		t.Error(err)
		return
	}

	resp, err := getResponse("http://" + socket + "/test")
	if err != nil {
		t.Error(err)
		return
	}

	if resp != "xyz456" {
		t.Errorf("invalid response received: %s", resp)
		return
	}

	err = server.Stop()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestCustomViewWithDynamicContent(t *testing.T) {
	server := getNewServer(t)
	if server == nil {
		return
	}

	handleErrorChannel(t, server.ErrorChan)

	err := server.AddView("test", fmt.Sprintf("today's date: <!--now:%s-->", time.RFC822))
	if err != nil {
		t.Error(err)
		return
	}

	err = server.Start()
	if err != nil {
		t.Error(err)
		return
	}

	resp, err := getResponse("http://" + socket + "/test")
	if err != nil {
		t.Error(err)
		return
	}

	if resp != fmt.Sprintf("today's date: %s", time.Now().UTC().Format(time.RFC822)) {
		t.Errorf("invalid response received: %s", resp)
		return
	}

	err = server.Stop()
	if err != nil {
		t.Error(err)
		return
	}
}

// NOTE THIS TEST REQUIRES MANUAL INTERVENTION!
func TestCustomForm(t *testing.T) {
	server := getNewServer(t)
	if server == nil {
		return
	}

	handleErrorChannel(t, server.ErrorChan)

	formHtml := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset=utf-8 />
    <title>Test Form</title>
    <!--gasp_js-->
</head>
<body onload="GASP.init('<!--server_socket-->')">
    <label><!--now--></label>
    <br/><br/>
	<label class="glabel">Enter 'test': </label>
    <input type="text" class="gtextbox" />
    <br/><br/>
    <button class="gbutton">Run Test</button>
    <br/><br/>
    <label id="success-message" class="glabel" style="visibility: hidden">Success!</label>
</body>
</html>`

	err := server.AddView("test", formHtml)
	if err != nil {
		t.Error(err)
		return
	}

	buttonClicked := false
	server.AddEventHandler("test", "gbutton0", "click", func(event *ui.ClientEvent) {
		textboxText := event.State.Textboxes[0].Text
		if textboxText != "test" {
			t.Errorf("textbox does not have the expected text (\"test\"): %s", textboxText)
			return
		}

		buttonText := event.State.Buttons[0].Text
		if buttonText != "Run Test" {
			t.Errorf("button does not have the expected text (\"Run Test\"): %s", buttonText)
			return
		}

		event.State.GetLabel("success-message").IsVisible = true
		event.State.Textboxes[0].IsEnabled = false
		event.State.Buttons[0].IsEnabled = false

		evt := ui.ServerEvent{
			Type: "form_update",
			Data: map[string]interface{}{"state": event.State},
		}
		server.SendEvent(&evt)

		t.Logf("button clicked: %s", buttonText)
		buttonClicked = true
	})

	err = server.Start()
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("ACCESS THE UI WITHIN 10 SECONDS TO COMPLETE THE TEST: http://%s/test", socket)
	time.Sleep(10 * time.Second)

	err = server.Stop()
	if err != nil {
		t.Error(err)
		return
	}

	if !buttonClicked {
		t.Error("button was not clicked or a failure occurred")
		return
	}
}

// NOTE THIS TEST REQUIRES MANUAL INTERVENTION!
func TestGaspForm1(t *testing.T) {
	buttonClicked := false

	form := ui.NewForm(ui.FormOptions{Socket: socket, Path: "test"}).
		AddVariableLabel("now").
		AddTextbox("Enter 'test': ").
		AddButton("Run Test", func(event *ui.ClientEvent) {
			textboxText := event.State.Textboxes[0].Text
			if textboxText != "test" {
				t.Errorf("textbox does not have the expected text (\"test\"): %s", textboxText)
				return
			}

			buttonText := event.State.Buttons[0].Text
			if buttonText != "Run Test" {
				t.Errorf("button does not have the expected text (\"Run Test\"): %s", buttonText)
				return
			}

			event.State.GetLabel("success-message").IsVisible = true
			event.State.Textboxes[0].IsEnabled = false
			event.State.Buttons[0].IsEnabled = false
			event.Form.Update(&event.State)

			t.Logf("button clicked: %s", buttonText)
			buttonClicked = true
		}).
		AddLabel("Success!", ui.ControlAttribute{Key: "id", Value: "success-message"}, ui.ControlAttribute{Key: "style", Value: "visibility:hidden"})

	handleErrorChannel(t, form.ErrorChan)

	_, err := form.Start()
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("ACCESS THE UI WITHIN 10 SECONDS TO COMPLETE THE TEST: %s", form.GetUri())
	time.Sleep(10 * time.Second)

	err = form.Stop()
	if err != nil {
		t.Error(err)
		return
	}

	if !buttonClicked {
		t.Error("button was not clicked or a failure occurred")
		return
	}
}

// NOTE THIS TEST REQUIRES MANUAL INTERVENTION!
func TestGaspForm2(t *testing.T) {
	buttonClicked := false

	form := ui.NewForm().
		AddDropdown("Select 'bird': ", []string{"cat", "dog", "bird", "fish"}).
		AddCheckbox("Check me!").
		AddButton("Run Test", func(event *ui.ClientEvent) {
			pet := event.State.Dropdowns[0].Text
			if pet != "bird" {
				t.Errorf("dropdown does not have the expected text (\"bird\"): %s", pet)
				return
			}

			if !event.State.Checkboxes[0].IsChecked {
				t.Error("checkbox was not selected or a failure occurred")
				return
			}

			event.State.Dropdowns[0].Text = "dog"
			event.State.Checkboxes[0].IsChecked = false
			event.Form.Update(&event.State)

			buttonClicked = true
			t.Log("button clicked")
		})

	handleErrorChannel(t, form.ErrorChan)

	_, err := form.Start()
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("ACCESS THE UI WITHIN 10 SECONDS TO COMPLETE THE TEST: %s", form.GetUri())
	time.Sleep(10 * time.Second)

	err = form.Stop()
	if err != nil {
		t.Error(err)
		return
	}

	if !buttonClicked {
		t.Error("button was not clicked or a failure occurred")
		return
	}
}

// NOTE THIS TEST REQUIRES MANUAL INTERVENTION!
func TestGaspForm3(t *testing.T) {
	button1Clicked := false
	button2Clicked := false

	form := ui.NewForm(ui.FormOptions{Socket: socket, Path: "test"}).
		AddLabel("After you click the 'Run Test' button...").
		AddTextbox("...the word 'cat' should appear in this textbox:").
		AddTextbox("...this textbox should get disabled:").
		AddTextbox("...this textbox should disappear:").
		AddLabel("Success!", ui.ControlAttribute{Key: "id", Value: "success-message"}, ui.ControlAttribute{Key: "style", Value: "visibility:hidden"}).
		AddButton("Run Test", func(event *ui.ClientEvent) {
			event.Form.GetTextbox("gtextbox0").UpdateText("cat")
			event.Form.GetTextbox("gtextbox1").UpdateIsEnabled(false)
			event.Form.GetTextbox("gtextbox2").UpdateIsVisible(false)

			t.Logf("button 1 clicked")
			button1Clicked = true
		}).
		AddButton("Check State", func(event *ui.ClientEvent) {
			if event.State.Textboxes[0].Text != "cat" {
				t.Errorf("textbox does not have the expected text (\"cat\"): %s", event.State.Textboxes[0].Text)
				return
			}

			if event.State.Textboxes[1].IsEnabled {
				t.Error("textbox is expected to be disabled")
				return
			}

			if event.State.Textboxes[2].IsVisible {
				t.Error("textbox is expected to be hidden")
				return
			}

			event.Form.GetLabel("success-message").UpdateIsVisible(true)

			t.Logf("button 2 clicked")
			button2Clicked = true
		})

	handleErrorChannel(t, form.ErrorChan)

	_, err := form.Start()
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("ACCESS THE UI WITHIN 10 SECONDS TO COMPLETE THE TEST: %s", form.GetUri())
	time.Sleep(10 * time.Second)

	err = form.Stop()
	if err != nil {
		t.Error(err)
		return
	}

	if !button1Clicked {
		t.Error("the 'Run Test' button was not clicked or a failure occurred")
		return
	}

	if !button2Clicked {
		t.Error("the 'Check State' button was not clicked or a failure occurred")
		return
	}
}

// NOTE THIS TEST REQUIRES MANUAL INTERVENTION!
func TestGaspForm4(t *testing.T) {
	var formLayoutIsCorrect *bool

	form, err := ui.NewForm(ui.FormOptions{Socket: socket, Path: "test"}).
		AddTextbox("First Name").
		AddTextbox("Address").
		AddTextbox("City").
		AddButton("Looks Bad!", func(event *ui.ClientEvent) {
			isCorrect := false
			formLayoutIsCorrect = &isCorrect
			t.Log("the layout failed validation")
		}).
		AddColumn().
		AddTextbox("Last Name").
		AddTextbox("Suite").
		AddTextbox("State").
		AddButton("Looks Good!", func(event *ui.ClientEvent) {
			isCorrect := true
			formLayoutIsCorrect = &isCorrect
			t.Log("the layout passed validation")
		}).
		Start()

	if err != nil {
		t.Error(err)
		return
	}

	handleErrorChannel(t, form.ErrorChan)

	t.Logf("ACCESS THE UI WITHIN 10 SECONDS TO COMPLETE THE TEST: %s", form.GetUri())
	time.Sleep(10 * time.Second)

	err = form.Stop()
	if err != nil {
		t.Error(err)
		return
	}

	if formLayoutIsCorrect == nil {
		t.Error("the form layout was not reviewed, no button was clicked")
		return
	}

	if !*formLayoutIsCorrect {
		t.Error("the form layout has an issue")
		return
	}
}

// NOTE THIS TEST REQUIRES MANUAL INTERVENTION!
func TestLineChart1(t *testing.T) {
	addPoint := func(event *ui.ClientEvent, thetaVar string, line string, increment float64) {
		theta := float64(0)
		if lastTheta, ok := event.Form.Data[thetaVar].(float64); ok {
			theta = lastTheta
		}
		theta += increment
		event.Form.Data[thetaVar] = theta
		y := math.Sin(theta) * 512
		event.Form.GetLineChart("glinechart0").UpdateValues(line, y)
	}

	form, err := ui.NewForm().
		AddButton("Add Point", func(event *ui.ClientEvent) {
			addPoint(event, "theta0", "line0", 0.1)
		}).
		AddButton("Start Signal", func(event *ui.ClientEvent) {
			if event.Form.Data["abort_chan0"] != nil {
				return
			}
			abortChan := make(chan bool)
			event.Form.Data["abort_chan0"] = abortChan

			go func() {
				for {
					select {
					case <-abortChan:
						event.Form.Data["abort_chan0"] = nil
						return
					default:
						addPoint(event, "theta0", "line0", 0.1)
						time.Sleep(15 * time.Millisecond)
					}

				}
			}()
		}).
		AddButton("Stop Signal", func(event *ui.ClientEvent) {
			if abortChan, ok := event.Form.Data["abort_chan0"].(chan bool); ok {
				abortChan <- true
			}
		}).
		AddColumn().
		AddLineChart(ui.LineChartState{
			ControlState: ui.ControlState{Text: "Sine Wave"},
			Width:        1000, Height: 200,
			Lines: []*ui.LineState{
				{
					Thickness: 2,
					Color:     "blue",
				},
				{
					Thickness: 2,
					Color:     "red",
				},
			},
		}, ui.ControlAttribute{Key: "style", Value: "margin:auto"}).
		AddColumn().
		AddButton("Add Point", func(event *ui.ClientEvent) {
			addPoint(event, "theta1", "line1", 0.03)
		}).
		AddButton("Start Signal", func(event *ui.ClientEvent) {
			if event.Form.Data["abort_chan1"] != nil {
				return
			}
			abortChan := make(chan bool)
			event.Form.Data["abort_chan1"] = abortChan
			go func() {
				for {
					select {
					case <-abortChan:
						event.Form.Data["abort_chan1"] = nil
						return
					default:
						addPoint(event, "theta1", "line1", 0.03)
						time.Sleep(15 * time.Millisecond)
					}

				}
			}()
		}).
		AddButton("Stop Signal", func(event *ui.ClientEvent) {
			if abortChan, ok := event.Form.Data["abort_chan1"].(chan bool); ok {
				abortChan <- true
			}
		}).
		Start()

	if err != nil {
		t.Error(err)
		return
	}

	handleErrorChannel(t, form.ErrorChan)

	t.Logf("ACCESS THE UI WITHIN 60 SECONDS TO COMPLETE THE TEST: %s", form.GetUri())
	time.Sleep(60 * time.Second)

	err = form.Stop()
	if err != nil {
		t.Error(err)
		return
	}
}

// NOTE THIS TEST REQUIRES MANUAL INTERVENTION!
func TestLineChart2(t *testing.T) {
	form, err := ui.NewForm().
		AddButton("Start Signal", func(event *ui.ClientEvent) {
			if event.Form.Data["abort_chan"] != nil {
				return
			}
			abortChan := make(chan bool)
			event.Form.Data["abort_chan"] = abortChan
			go func() {
				for {
					select {
					case <-abortChan:
						event.Form.Data["abort_chan"] = nil
						return
					default:
						event.Form.GetLineChart("glinechart0").UpdateValues("line0", rand.Float64())
						time.Sleep(20 * time.Millisecond)
					}

				}
			}()
		}).
		AddButton("Stop Signal", func(event *ui.ClientEvent) {
			if abortChan, ok := event.Form.Data["abort_chan"].(chan bool); ok {
				abortChan <- true
			}
		}).
		AddButton("Change Styling", func(event *ui.ClientEvent) {
			event.State.LineCharts[0].Text = ""
			event.State.LineCharts[0].Lines[0].Thickness = 5
			event.State.LineCharts[0].Lines[0].Color = "green"
			event.Form.Update(&event.State)
		}).
		AddColumn().
		AddLineChart(ui.LineChartState{
			ControlState: ui.ControlState{Text: "Noise"},
			Width:        1000, Height: 200,
			Lines: []*ui.LineState{
				{
					Thickness: 1,
					Color:     "blue",
				},
			},
		}, ui.ControlAttribute{Key: "data-buffer-size", Value: "250"}, ui.ControlAttribute{Key: "data-axes-thickness", Value: "0"}).
		Start()

	if err != nil {
		t.Error(err)
		return
	}

	handleErrorChannel(t, form.ErrorChan)

	t.Logf("ACCESS THE UI WITHIN 60 SECONDS TO COMPLETE THE TEST: %s", form.GetUri())
	time.Sleep(60 * time.Second)

	err = form.Stop()
	if err != nil {
		t.Error(err)
		return
	}
}

// NOTE THIS TEST REQUIRES MANUAL INTERVENTION!
func TestPacketInspector(t *testing.T) {
	passed := false
	form, err := ui.NewForm().
		AddPacketInspector(ui.PacketInspectorState{}).
		AddButton("Load Packet", func(event *ui.ClientEvent) {
			testValue := byte(133) // signed = -123
			event.Form.GetPacketInspector("gpacketinspector0").UpdateBytes([]byte{testValue})
		}).
		AddEventHandler("*", "gpacketinspector0", "selection_update", func(event *ui.ClientEvent) {
			if uint(event.Data["uint"].(float64)) == 133 && int(event.Data["int"].(float64)) == -123 {
				passed = true
			}
		}).
		Start()

	if err != nil {
		t.Error(err)
		return
	}

	handleErrorChannel(t, form.ErrorChan)

	t.Log("Click the 'Load Packet' button and select all eight bits, from least significant (right-most) to most significant (left-most).")

	t.Logf("ACCESS THE UI WITHIN 20 SECONDS TO COMPLETE THE TEST: %s", form.GetUri())
	time.Sleep(20 * time.Second)

	if !passed {
		t.Error("selected integer value is not as expected or the test was not performed correctly")
	}

	err = form.Stop()
	if err != nil {
		t.Error(err)
		return
	}
}

// NOTE THIS TEST REQUIRES MANUAL INTERVENTION!
func TestPacketInspectorWithLineChart(t *testing.T) {
	form, err := ui.NewForm().
		AddLabel("There is a 10-bit sine wave embedded within the noise, starting from the 8th least significant bit...").
		AddLineChart(ui.LineChartState{
			Width: 1000, Height: 200,
			Lines: []*ui.LineState{
				{
					Thickness: 2,
					Color:     "green",
				},
			},
		}).
		AddPacketInspector(ui.PacketInspectorState{
			Width:     1000,
			ByteCount: 4,
		}).
		AddButton("Start Signal", func(event *ui.ClientEvent) {
			if event.Form.Data["abort_chan"] != nil {
				return
			}
			abortChan := make(chan bool)
			event.Form.Data["abort_chan"] = abortChan
			go func() {
				theta := 0.0

				for {
					theta += 0.01
					select {
					case <-abortChan:
						event.Form.Data["abort_chan"] = nil
						return
					default:
						r := uint32(rand.Int())
						r = 0b11111111111111100000000001111111 & r
						y := uint32((math.Sin(theta)*512)+512) << 7
						y = 0b00000000000000011111111110000000 & y
						z := r | y
						bytes := make([]byte, 4)
						binary.BigEndian.PutUint32(bytes, z)
						event.Form.GetPacketInspector("gpacketinspector0").UpdateBytes(bytes)
						time.Sleep(20 * time.Millisecond)
					}

				}
			}()
		}).
		AddButton("Stop Signal", func(event *ui.ClientEvent) {
			if abortChan, ok := event.Form.Data["abort_chan"].(chan bool); ok {
				abortChan <- true
			}
		}).
		AddEventHandler("*", "gpacketinspector0", "selection_update", func(event *ui.ClientEvent) {
			value := event.Data["uint"].(float64)
			event.Form.GetLineChart("glinechart0").UpdateValues("line0", value)
		}).
		Start()

	if err != nil {
		t.Error(err)
		return
	}

	handleErrorChannel(t, form.ErrorChan)

	t.Logf("ACCESS THE UI WITHIN 60 SECONDS TO COMPLETE THE TEST: %s", form.GetUri())
	time.Sleep(60 * time.Second)

	err = form.Stop()
	if err != nil {
		t.Error(err)
		return
	}
}

// NOTE THIS TEST REQUIRES MANUAL INTERVENTION!
func TestCustomEvents(t *testing.T) {
	server := getNewServer(t)
	if server == nil {
		return
	}

	handleErrorChannel(t, server.ErrorChan)

	formHtml := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset=utf-8 />
    <title>Test Form</title>
    <!--gasp_css-->
	<!--gasp_js-->
</head>
<body onload="GASP.init('<!--server_socket-->'); GASP.addServerEventHandler('pong', receivePong);">
	<div class="gform">
		<button class="gbutton" onclick="sendPing(this.id, this.innerHTML);">Rick</button>
		<br/><br/>
		<button class="gbutton" onclick="sendPing(this.id, this.innerHTML);">Morty</button>
		<br/><br/>
		<label id="message" class="glabel" style="visibility: hidden">_</label>
	</div>
	<script>
		function sendPing(id, name) {
			GASP.sendEvent(GASP.newEvent(id, 'ping', { 'name': name } ));
		}

		function receivePong(evt) {
			let msgLabel = document.getElementById('message');
			msgLabel.innerHTML = evt.text;
			msgLabel.style.visibility = 'visible';
		}
	</script>
</body>
</html>`

	err := server.AddView("/", formHtml)
	if err != nil {
		t.Error(err)
		return
	}

	pingReceived := false

	// Handle the custom 'ping' event from any (*) view and any element
	server.AddEventHandler("*", "*", "ping", func(event *ui.ClientEvent) {
		pingReceived = true
		t.Logf("ping received from button '%s'", event.Id)

		if name, ok := event.Data["name"].(string); ok {
			server.SendEvent(&ui.ServerEvent{
				Type: "pong",
				Text: "Hello " + name + "!",
			})
		} else {
			t.Error("name not received from client ping")
			return
		}
	})

	// Handle the custom 'ping' event from any (*) view but only from 'gbutton0'
	server.AddEventHandler("*", "gbutton0", "ping", func(event *ui.ClientEvent) {
		if name, ok := event.Data["name"].(string); ok {
			if name != "Rick" {
				t.Error("expected ping from Rick")
				return
			}
			t.Logf("ping received from Rick")
		} else {
			t.Error("name not received from client ping")
			return
		}
	})

	// Handle the custom 'ping' event from any (*) view but only from 'gbutton1'
	server.AddEventHandler("*", "gbutton1", "ping", func(event *ui.ClientEvent) {
		if name, ok := event.Data["name"].(string); ok {
			if name != "Morty" {
				t.Error("expected ping from Morty")
				return
			}
			t.Logf("ping received from Morty")
		} else {
			t.Error("name not received from client ping")
			return
		}
	})

	err = server.Start()
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("ACCESS THE UI WITHIN 10 SECONDS TO COMPLETE THE TEST: http://%s/test", socket)
	time.Sleep(10 * time.Second)

	err = server.Stop()
	if err != nil {
		t.Error(err)
		return
	}

	if !pingReceived {
		t.Error("ping was not received")
		return
	}
}

func TestRouteGuards(t *testing.T) {
	server := getNewServer(t)
	if server == nil {
		return
	}

	handleErrorChannel(t, server.ErrorChan)

	err := server.AddView("secured", "access granted")
	if err != nil {
		t.Error(err)
		return
	}

	err = server.AddView("denied", "access denied")
	if err != nil {
		t.Error(err)
		return
	}

	err = server.AddRouteGuard("secured", func(req *http.Request) *string {
		password := req.URL.Query().Get("pw")
		if password == "mySecret" {
			return nil
		} else {
			newPath := "denied"
			return &newPath
		}
	})

	err = server.Start()
	if err != nil {
		t.Error(err)
		return
	}

	resp, err := getResponse("http://" + socket + "/secured?pw=mySecret")
	if err != nil {
		t.Error(err)
		return
	}

	if resp != "access granted" {
		t.Errorf("invalid response received: %s", resp)
		return
	}

	resp, err = getResponse("http://" + socket + "/secured?pw=xxx")
	if err != nil {
		t.Error(err)
		return
	}

	if resp != "access denied" {
		t.Errorf("invalid response received: %s", resp)
		return
	}

	err = server.Stop()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestResources(t *testing.T) {
	server := getNewServer(t)
	if server == nil {
		return
	}

	handleErrorChannel(t, server.ErrorChan)

	err := server.AddResources("resources", ".go", ".sh", ".crt", ".key")
	if err != nil {
		t.Error(err)
		return
	}

	err = server.Start()
	if err != nil {
		t.Error(err)
		return
	}

	resp, err := getResponse("http://" + socket + "/css/test.css")
	if err != nil {
		t.Error(err)
		return
	}

	if !strings.HasPrefix(resp, ".testClass{}") {
		t.Errorf("invalid response received: %s", resp)
		return
	}

	resp, err = getResponse("http://" + socket + "/certs/test-ca.key")
	if err == nil {
		t.Error("expected not to be able to download test-ca.key")
		return
	}

	err = server.Stop()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestVariables1(t *testing.T) {
	server := getNewServer(t)
	if server == nil {
		return
	}

	handleErrorChannel(t, server.ErrorChan)

	hostname, _ := os.Hostname()

	err := server.AddView("", "<!--hostname-->")
	if err != nil {
		t.Error(err)
		return
	}

	err = server.AddVariableSetter("hostname", func(req *http.Request) string {
		return hostname
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = server.Start()
	if err != nil {
		t.Error(err)
		return
	}

	resp, err := getResponse("http://" + socket + "/")
	if err != nil {
		t.Error(err)
		return
	}

	if resp != hostname {
		t.Errorf("invalid response received: %s", resp)
		return
	}

	err = server.Stop()
	if err != nil {
		t.Error(err)
		return
	}
}

// NOTE THIS TEST REQUIRES MANUAL INTERVENTION!
func TestVariables2(t *testing.T) {
	hostname, err := os.Hostname()
	if err != nil {
		t.Error(err)
		return
	}

	buttonClicked := false

	form := ui.NewForm().
		AddVariableSetter("hostname", func(req *http.Request) string {
			return hostname
		}).
		AddVariableLabel("hostname").
		AddButton("Run Test", func(event *ui.ClientEvent) {
			t.Log("button clicked")

			labelText := event.State.Labels[0].Text
			if labelText != hostname {
				t.Error(fmt.Errorf("label text '%s' does not match expected value ('%s')", labelText, hostname))
				return
			}
			buttonClicked = true
		})

	handleErrorChannel(t, form.ErrorChan)

	_, err = form.Start()
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("ACCESS THE UI WITHIN 10 SECONDS TO COMPLETE THE TEST: http://%s", socket)
	time.Sleep(10 * time.Second)

	err = form.Stop()
	if err != nil {
		t.Error(err)
		return
	}

	if !buttonClicked {
		t.Error("button was not clicked or a failure occurred")
		return
	}
}

// NOTE THIS TEST REQUIRES MANUAL INTERVENTION!
func TestTLS(t *testing.T) {
	buttonClicked := false

	form := ui.NewForm().
		AddButton("Run Test", func(event *ui.ClientEvent) {
			t.Log("button clicked")
			buttonClicked = true
		})

	handleErrorChannel(t, form.ErrorChan)

	err := form.StartWithTLS("resources/certs/test-server.crt", "resources/certs/test-server.key")
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("ACCESS THE UI WITHIN 10 SECONDS TO COMPLETE THE TEST: %s", form.GetUri())
	time.Sleep(10 * time.Second)

	err = form.Stop()
	if err != nil {
		t.Error(err)
		return
	}

	if !buttonClicked {
		t.Error("button was not clicked or a failure occurred")
		return
	}
}

func getNewServer(t *testing.T) *ui.Server {
	server, err := ui.NewServer(socket)
	if err != nil {
		t.Error(err)
		return nil
	}
	return server
}

func handleErrorChannel(t *testing.T, errorChan chan error) {
	go func() {
		for {
			err := <-errorChan
			if err != nil && !errors.Is(err, http.ErrServerClosed) && !websocket.IsCloseError(err, websocket.CloseGoingAway) {
				t.Error(err)
				return
			}
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
		}
	}()
}

func getResponse(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid status code received: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
