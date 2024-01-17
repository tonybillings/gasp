package ui

import (
	"fmt"
	ui "tonysoft.com/gasp"
)

var (
	form                *ui.Form
	lineChartUpdateChan chan *ui.LineChartState
)

func Init(socket string) {
	formOptions := ui.FormOptions{
		Socket: socket,
	}
	form = ui.NewForm(formOptions)
	lineChartUpdateChan = make(chan *ui.LineChartState, 100)

	go lineChartUpdateRoutine()

	fmt.Printf("ACCESS GOWATCH UI VIA: %s\n\n", form.GetUri())
}

func AddLineChart(initialState ...ui.LineChartState) {
	if initialState != nil && len(initialState) > 0 {
		form.AddLineChart(initialState[0])
	} else {
		form.AddLineChart(ui.LineChartState{})
	}
}

func UpdateLineChart(id string, value float64) {
	state := ui.LineChartState{}
	state.Id = id
	state.Lines = []*ui.LineState{{NewValues: []float64{value}}}
	lineChartUpdateChan <- &state
}

func Start() {
	_, err := form.Start()
	if err != nil {
		panic(err)
	}
}

func Stop() {
	err := form.Stop()
	if err != nil {
		panic(err)
	}

	close(lineChartUpdateChan)
}

func lineChartUpdateRoutine() {
	for {
		state, ok := <-lineChartUpdateChan
		if !ok {
			return
		}
		form.GetLineChart(state.Id).UpdateValues("", state.Lines[0].NewValues...)
	}
}
