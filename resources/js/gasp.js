const GASP_proto = {
    getViewName() {
        return window.location.pathname.substring(1);
    },
    getControlStates() {
        let state = {};
        state.buttons = [];
        for (let i = 0; i < this.buttons.length; i++) {
            let btn = this.buttons[i];
            let btnState = {};
            btnState.id = btn.id;
            btnState.text = btn.textContent;
            btnState.is_visible = this.getControlIsVisible(btn);
            btnState.is_enabled = this.getControlIsEnabled(btn);
            state.buttons.push(btnState);
        }
        state.textboxes = [];
        for (let i = 0; i < this.textboxes.length; i++) {
            let tb = this.textboxes[i];
            let tbState = {};
            tbState.id = tb.id;
            tbState.text = tb.value;
            tbState.is_visible = this.getControlIsVisible(tb);
            tbState.is_enabled = this.getControlIsEnabled(tb);
            state.textboxes.push(tbState);
        }
        state.labels = [];
        for (let i = 0; i < this.labels.length; i++) {
            let lbl = this.labels[i];
            let lblState = {};
            lblState.id = lbl.id;
            lblState.text = lbl.innerHTML;
            lblState.is_visible = this.getControlIsVisible(lbl);
            lblState.is_enabled = this.getControlIsEnabled(lbl);
            state.labels.push(lblState);
        }
        state.dropdowns = [];
        for (let i = 0; i < this.dropdowns.length; i++) {
            let dd = this.dropdowns[i];
            let ddState = {};
            ddState.id = dd.id;
            ddState.text = dd.options[dd.selectedIndex].text;
            ddState.is_visible = this.getControlIsVisible(dd);
            ddState.is_enabled = this.getControlIsEnabled(dd);
            state.dropdowns.push(ddState);
        }
        state.checkboxes = [];
        for (let i = 0; i < this.checkboxes.length; i++) {
            let cb = this.checkboxes[i];
            let cbState = {};
            cbState.id = cb.id;
            cbState.is_checked = cb.checked;
            cbState.is_visible = this.getControlIsVisible(cb);
            cbState.is_enabled = this.getControlIsEnabled(cb);
            state.checkboxes.push(cbState);
        }
        state.linecharts = [];
        for (let i = 0; i < this.linecharts.length; i++) {
            let lc = this.linecharts[i].linechart;
            let lcState = {};
            lcState.id = lc.id;
            lcState.is_visible = this.getControlIsVisible(lc.canvas);
            lcState.is_enabled = this.getControlIsEnabled(lc.canvas);

            lcState.lines = [];
            for (let [_lineName, line] of lc.lines) {
                lcState.lines.push(line);
            }

            state.linecharts.push(lcState);
        }
        state.packetinspectors = [];
        for (let i = 0; i < this.packetinspectors.length; i++) {
            let pi = this.packetinspectors[i].packetinspector;
            let piState = {};
            piState.id = pi.id;
            piState.is_visible = this.getControlIsVisible(pi.canvas);
            piState.is_enabled = this.getControlIsEnabled(pi.canvas);
            state.packetinspectors.push(piState);
        }
        return state;
    },
    getControlIsVisible(control) {
        return control.style.visibility === '' || control.style.visibility === 'visible';
    },
    getControlIsEnabled(control) {
        return control.getAttribute('disabled') == null;
    },
    setControlIsVisible(control, isVisible) {
        if (control.style) {
            control.style.visibility = isVisible ? 'visible' : 'hidden';
        }
    },
    setControlIsEnabled(control, isEnabled) {
        if (control.setAttribute) {
            if (isEnabled) {
                control.removeAttribute('disabled');
            } else {
                control.setAttribute('disabled', 'disabled');
            }
        }
    },
    updateFormState(state) {
        state.textboxes.forEach(control => {
            let ctl = document.getElementById(control.id);
            ctl.value = control.text;
            this.setControlIsVisible(ctl, control.is_visible);
            this.setControlIsEnabled(ctl, control.is_enabled);
        });

        state.buttons.forEach(control => {
            let ctl = document.getElementById(control.id);
            ctl.textContent = control.text;
            this.setControlIsVisible(ctl, control.is_visible);
            this.setControlIsEnabled(ctl, control.is_enabled);
        });

        state.labels.forEach(control => {
            let ctl = document.getElementById(control.id);
            ctl.innerHTML = control.text;
            this.setControlIsVisible(ctl, control.is_visible);
            this.setControlIsEnabled(ctl, control.is_enabled);
        });

        state.dropdowns.forEach(control => {
            let ctl = document.getElementById(control.id);
            ctl.selectedIndex = -1;
            for (let i = 0; i < ctl.options.length; i++) {
                if (ctl.options[i].text === control.text) {
                    ctl.selectedIndex = i;
                    break;
                }
            }
            this.setControlIsVisible(ctl, control.is_visible);
            this.setControlIsEnabled(ctl, control.is_enabled);
        });

        state.checkboxes.forEach(control => {
            let ctl = document.getElementById(control.id);
            ctl.checked = control.is_checked;
            this.setControlIsVisible(ctl, control.is_visible);
            this.setControlIsEnabled(ctl, control.is_enabled);
        });

        state.linecharts.forEach(control => {
            let lineChart = null;
            for (let i = 0; i < this.linecharts.length; i++) {
                let canvas = this.linecharts[i];
                if (canvas.id === control.id) {
                    lineChart = canvas.linechart;
                    break;
                }
            }

            for (let i = 0; i < control.lines.length; i++) {
                let lineState = control.lines[i];
                let line = lineChart.lines.get(lineState.name);
                line.thickness = lineState.thickness;
                line.color = lineState.color;
            }

            lineChart.text = control.text;
            this.setControlIsVisible(lineChart, control.is_visible);
            this.setControlIsEnabled(lineChart, control.is_enabled);
        });

        state.packetinspectors.forEach(control => {
            let packetinspector = null;
            for (let i = 0; i < this.linecharts.length; i++) {
                let canvas = this.linecharts[i];
                if (canvas.id === control.id) {
                    packetinspector = canvas.packetinspector;
                    break;
                }
            }

            packetinspector.text = control.text;
            this.setControlIsVisible(packetinspector, control.is_visible);
            this.setControlIsEnabled(packetinspector, control.is_enabled);
        });
    },
    updateTextbox(data) {
        let ctl = document.getElementById(data.state.id);
        for (let i = 0; i < data.properties.length; i++) {
            switch (data.properties[i]) {
                case 'text':
                    ctl.value = data.state.text;
                    break;
                case 'is_visible':
                    this.setControlIsVisible(ctl, data.state.is_visible);
                    break;
                case 'is_enabled':
                    this.setControlIsEnabled(ctl, data.state.is_enabled);
                    break;
            }
        }
    },
    updateButton(data) {
        let ctl = document.getElementById(data.state.id);
        for (let i = 0; i < data.properties.length; i++) {
            switch (data.properties[i]) {
                case 'text':
                    ctl.textContent = data.state.text;
                    break;
                case 'is_visible':
                    this.setControlIsVisible(ctl, data.state.is_visible);
                    break;
                case 'is_enabled':
                    this.setControlIsEnabled(ctl, data.state.is_enabled);
                    break;
            }
        }
    },
    updateLabel(data) {
        let ctl = document.getElementById(data.state.id);
        for (let i = 0; i < data.properties.length; i++) {
            switch (data.properties[i]) {
                case 'text':
                    ctl.innerHTML = data.state.text;
                    break;
                case 'is_visible':
                    this.setControlIsVisible(ctl, data.state.is_visible);
                    break;
                case 'is_enabled':
                    this.setControlIsEnabled(ctl, data.state.is_enabled);
                    break;
            }
        }
    },
    updateDropdown(data) {
        let ctl = document.getElementById(data.state.id);
        for (let i = 0; i < data.properties.length; i++) {
            switch (data.properties[i]) {
                case 'text':
                    ctl.selectedIndex = -1;
                    for (let i = 0; i < ctl.options.length; i++) {
                        if (ctl.options[i].text === data.state.text) {
                            ctl.selectedIndex = i;
                            break;
                        }
                    }
                    break;
                case 'is_visible':
                    this.setControlIsVisible(ctl, data.state.is_visible);
                    break;
                case 'is_enabled':
                    this.setControlIsEnabled(ctl, data.state.is_enabled);
                    break;
            }
        }
    },
    updateCheckbox(data) {
        let ctl = document.getElementById(data.state.id);
        for (let i = 0; i < data.properties.length; i++) {
            switch (data.properties[i]) {
                case 'is_checked':
                    ctl.checked = data.state.is_checked;
                    break;
                case 'is_visible':
                    this.setControlIsVisible(ctl, data.state.is_visible);
                    break;
                case 'is_enabled':
                    this.setControlIsEnabled(ctl, data.state.is_enabled);
                    break;
            }
        }
    },
    updateLineChart(data) {
        let lineChart = null;
        for (let i = 0; i < this.linecharts.length; i++) {
            let canvas = this.linecharts[i];
            if (canvas.id === data.state.id) {
                lineChart = canvas.linechart;
                break;
            }
        }
        for (let i = 0; i < data.properties.length; i++) {
            switch (data.properties[i]) {
                case 'values':
                    for (let j = 0; j < data.state.lines.length; j++) {
                        let lineState = data.state.lines[j];
                        lineChart.addValues(lineState.name, lineState.new_values);
                    }
                    break;
                case 'lines':
                    for (let j = 0; j < data.state.lines.length; j++) {
                        let lineState = data.state.lines[j];
                        let line = lineChart.lines.get(line.name);
                        line.thickness = lineState.thickness;
                        line.color = lineState.color;
                    }
                    break;
                case 'is_visible':
                    this.setControlIsVisible(ctl, data.state.is_visible);
                    break;
                case 'is_enabled':
                    this.setControlIsEnabled(ctl, data.state.is_enabled);
                    break;
            }
        }
    },
    updatePacketInspector(data) {
        let inspector = null;
        for (let i = 0; i < this.packetinspectors.length; i++) {
            let canvas = this.packetinspectors[i];
            if (canvas.id === data.state.id) {
                inspector = canvas.packetinspector;
                break;
            }
        }
        for (let i = 0; i < data.properties.length; i++) {
            switch (data.properties[i]) {
                case 'packet':
                    inspector.updateBytes(data.state.packet);
                    break;
                case 'is_visible':
                    this.setControlIsVisible(ctl, data.state.is_visible);
                    break;
                case 'is_enabled':
                    this.setControlIsEnabled(ctl, data.state.is_enabled);
                    break;
            }
        }
    },
    newEvent(id, type, data) {
        let evt = {};
        evt.view = this.getViewName();
        evt.id = id;
        evt.type = type;
        evt.data = data;
        evt.state = this.getControlStates();
        return evt;
    },
    sendEvent(evt) {
        this.socket.send(JSON.stringify(evt));
    },
    addControlEventHandler(id, eventType) {
        if (id === '*') {
            return;
        }
        let ctl = document.getElementById(id);
        if (!ctl) {
            return;
        }

        ctl.addEventListener(eventType, (evt) => {
            this.sendEvent(this.newEvent(id, evt.type))
        });
    },
    addServerEventHandler(eventType, func) {
        if (!this.serverEventHandlers) {
            this.serverEventHandlers = {};
        }

        if (!this.serverEventHandlers[eventType]) {
            this.serverEventHandlers[eventType] = [];
        }

        this.serverEventHandlers[eventType].push(func);
    },
    tick(timestamp) {
        if (this.linecharts) {
            for (let i = 0; i < this.linecharts.length; i++) {
                this.linecharts[i].linechart.tick(timestamp);
            }
        }

        if (this.packetinspectors) {
            for (let i = 0; i < this.packetinspectors.length; i++) {
                this.packetinspectors[i].packetinspector.tick(timestamp);
            }
        }
    },
    frameRequestCallback(timestamp) {
        GASP.tick(timestamp);
        window.requestAnimationFrame(GASP.frameRequestCallback);
    },
    initControls() {
        this.buttons = document.querySelectorAll('.gbutton');
        for (let i = 0; i < this.buttons.length; i++) {
            let id = this.buttons[i].id;
            if (!id || id === '') {
                this.buttons[i].id = 'gbutton' + i;
            }
        }

        this.textboxes = document.querySelectorAll('.gtextbox');
        for (let i = 0; i < this.textboxes.length; i++) {
            let id = this.textboxes[i].id;
            if (!id || id === '') {
                this.textboxes[i].id = 'gtextbox' + i;
            }
        }

        this.labels = document.querySelectorAll('.glabel');
        for (let i = 0; i < this.labels.length; i++) {
            let id = this.labels[i].id;
            if (!id || id === '') {
                this.labels[i].id = 'glabel' + i;
            }
        }

        this.dropdowns = document.querySelectorAll('.gdropdown');
        for (let i = 0; i < this.dropdowns.length; i++) {
            let id = this.dropdowns[i].id;
            if (!id || id === '') {
                this.dropdowns[i].id = 'gdropdown' + i;
            }
        }

        this.checkboxes = document.querySelectorAll('.gcheckbox');
        for (let i = 0; i < this.checkboxes.length; i++) {
            let id = this.checkboxes[i].id;
            if (!id || id === '') {
                this.checkboxes[i].id = 'gcheckbox' + i;
            }
        }

        let shouldRequestAnimationFrame = false;

        let lineChartCanvases = document.querySelectorAll('.glinechart');
        this.linecharts = [];
        for (let i = 0; i < lineChartCanvases.length; i++) {
            let canvas = lineChartCanvases[i];
            let id = canvas.id;
            if (!id || id === '') {
                id = 'glinechart' + i;
                canvas.id = id;
            }
            canvas.linechart = this.initLineChart(id);
            this.linecharts.push(canvas);
            shouldRequestAnimationFrame = true;
        }

        let packetInspectorCanvases = document.querySelectorAll('.gpacketinspector');
        this.packetinspectors = [];
        for (let i = 0; i < packetInspectorCanvases.length; i++) {
            let canvas = packetInspectorCanvases[i];
            let id = canvas.id;
            if (!id || id === '') {
                id = 'gpacketinspector' + i;
                canvas.id = id;
            }
            canvas.packetinspector = this.initPacketInspector(id);
            this.packetinspectors.push(canvas);
            shouldRequestAnimationFrame = true;
        }

        /*event_handlers*/

        if (shouldRequestAnimationFrame) {
            window.requestAnimationFrame(this.frameRequestCallback);
        }
    },
    initLineChart(id) {
        let ctl = document.getElementById(id);

        let initialState = JSON.parse(atob(ctl.dataset.initialState));

        let lineChart = {};
        lineChart.id = id;
        lineChart.text = initialState.text ?? '';
        lineChart.canvas = ctl;
        lineChart.context = lineChart.canvas.getContext('2d');
        lineChart.boundingRect = lineChart.context.canvas.getBoundingClientRect();
        lineChart.bufferSize = Number(ctl.dataset.bufferSize ?? lineChart.canvas.width);
        lineChart.minValue = Number(ctl.dataset.minValue ?? Number.MAX_VALUE);
        lineChart.maxValue = Number(ctl.dataset.maxValue ?? Number.MIN_SAFE_INTEGER);
        lineChart.backgroundColor = ctl.dataset.backgroundColor ?? 'black';
        lineChart.axesThickness = Number(ctl.dataset.axesThickness ?? 1);
        lineChart.axesColor = ctl.dataset.axesColor ?? 'rgb(128,128,128)';
        lineChart.legendColor = ctl.dataset.legendColor ?? 'white';
        lineChart.legendBackgroundColorStart = ctl.dataset.legendBackgroundColorStart ?? 'rgba(0, 0, 0, 1)';
        lineChart.legendBackgroundColorEnd = ctl.dataset.legendBackgroundColorEnd ?? 'rgba(0, 0, 0, 0)';
        lineChart.legendFontFamily = ctl.dataset.legendFontFamily ?? 'Arial';
        lineChart.legendFontSize = Number(ctl.dataset.legendFontSize ?? 20);
        lineChart.selectedValuesBackgroundColor = ctl.dataset.selectedValuesBackgroundColor ?? 'rgba(255, 255, 255, 0.25)';
        lineChart.selectedValuesTextColor = ctl.dataset.selectedValuesTextColor ?? 'white';
        lineChart.context.font = lineChart.legendFontSize + 'px ' + lineChart.legendFontFamily;
        lineChart.charWidth = lineChart.context.measureText('0').width;
        lineChart.charWidthTrippled = lineChart.charWidth * 3;
        lineChart.charHeight = lineChart.legendFontSize;
        lineChart.charHalfHeight = lineChart.charHeight / 2;
        lineChart.charQuarterHeight = lineChart.charHeight / 4;
        lineChart.charEigthHeight = lineChart.charHeight / 8;
        lineChart.isDirty = false;

        lineChart.lines = new Map();
        if (initialState.lines) {
            for (let i = 0; i < initialState.lines.length; i++) {
                let line = initialState.lines[i];
                if (!line.name || line.name === '') {
                    line.name = 'line' + i;
                }

                if (!line.thickness || line.thickness === 0) {
                    line.thickness = 1;
                }

                if (!line.color || line.color === '') {
                    line.color = 'white';
                }

                line.values = [];
                lineChart.lines.set(line.name, line);
            }
        }

        let ctx = lineChart.context;

        lineChart.resetScale = function () {
            lineChart.minValue = Number.MAX_VALUE;
            lineChart.maxValue = Number.MIN_VALUE;
        };

        lineChart.reset = function () {
            if (lineChart.lines) {
                for (let [_lineName, line] of lineChart.lines) {
                    line.values = [];
                }
            }

            lineChart.resetScale();
        };

        lineChart.addValues = function (lineName, values) {
            let line = lineChart.lines.get(lineName);
            line.values.push(...values);

            for (let val of values) {
                lineChart.minValue = (val < lineChart.minValue) ? val : lineChart.minValue;
                lineChart.maxValue = (val > lineChart.maxValue) ? val : lineChart.maxValue;
            }

            lineChart.isDirty = true;
            return lineChart;
        };

        lineChart.drawAxes = function () {
            if (lineChart.axesThickness === 0) {
                return;
            }

            const width = ctx.canvas.width;
            const height = ctx.canvas.height;

            ctx.save();
            ctx.beginPath();

            ctx.lineWidth = lineChart.axesThickness;
            ctx.strokeStyle = lineChart.axesColor;
            ctx.moveTo(0, height / 2);
            ctx.lineTo(width, height / 2);
            ctx.moveTo(width / 2, 0);
            ctx.lineTo(width / 2, height);

            ctx.stroke();
            ctx.restore();
        };

        lineChart.drawValues = function () {
            for (let [_lineName, line] of lineChart.lines) {
                if (line.values.length === 0) continue;

                const width = ctx.canvas.width;
                const height = ctx.canvas.height;
                const step = width / (line.values.length - 1);
                const range = lineChart.maxValue - lineChart.minValue;

                ctx.save();
                ctx.beginPath();

                ctx.lineWidth = line.thickness;
                ctx.strokeStyle = line.color;

                for (let i = 0; i < line.values.length; i++) {
                    const value = line.values[i];
                    let offsetVal = value;

                    if (lineChart.minValue !== 0) {
                        offsetVal = value - lineChart.minValue;
                    }

                    const pct = offsetVal / range;
                    const y = height + (pct * -height);

                    if (i === 0) {
                        ctx.moveTo(0, y);
                    } else {
                        ctx.lineTo(i * step, y);
                    }
                }

                ctx.stroke();
                ctx.restore();
            }
        };

        lineChart.drawLegend = function () {
            const width = ctx.canvas.width;
            const height = ctx.canvas.height;

            ctx.save();
            ctx.beginPath();

            let gradient = ctx.createLinearGradient(0, height / 2, width * 0.2, height / 2);
            gradient.addColorStop(0, lineChart.legendBackgroundColorStart);
            gradient.addColorStop(1, lineChart.legendBackgroundColorEnd);

            ctx.fillStyle = gradient;
            ctx.fillRect(0, 0, width * 0.2, height);

            ctx.fillStyle = lineChart.legendColor;

            ctx.fillText(lineChart.text, 4, (height * 0.5) + (lineChart.legendFontSize * 0.25));
            ctx.fillText(lineChart.maxValue.toFixed(6), 4, lineChart.legendFontSize);
            ctx.fillText(lineChart.minValue.toFixed(6), 4, height - 4);

            ctx.restore();
        };

        lineChart.drawSelectedValues = function () {
            if (!lineChart.shouldDrawSelectedValues) {
                return;
            }

            let pctX = lineChart.mousePosX / ctx.canvas.width;
            let selectedValues = [];
            let maxWidth = 0;
            for (let entry of lineChart.lines.entries()) {
                let line = entry[1];
                if (line.values.length === 0) {
                    continue;
                }
                let selVal = (line.values[(pctX * (line.values.length - 1) >>> 0)]).toString();
                let width = lineChart.charWidth * selVal.length;
                if (width > maxWidth) {
                    maxWidth = width;
                }
                selectedValues.push({ color: line.color, value: selVal });
            }

            ctx.save();
            ctx.fillStyle = lineChart.selectedValuesBackgroundColor;
            ctx.fillRect(lineChart.mousePosX - 10, lineChart.mousePosY - 10, (lineChart.charWidth * 5) + maxWidth + 20, (selectedValues.length * lineChart.legendFontSize) + 20);
            for (let i = 0; i < selectedValues.length; i++) {
                let selVal = selectedValues[i];
                ctx.fillStyle = selVal.color;
                ctx.fillRect(lineChart.mousePosX + 10, lineChart.mousePosY + (i * lineChart.charHeight) + lineChart.charQuarterHeight, lineChart.charWidthTrippled, lineChart.charHalfHeight);
                ctx.fillStyle = lineChart.selectedValuesTextColor;
                ctx.fillText(selVal.value, lineChart.mousePosX + lineChart.charWidthTrippled + 20, lineChart.mousePosY + (i * lineChart.charHeight) + (lineChart.charHeight - lineChart.charEigthHeight));
            }

            ctx.restore();
        };

        lineChart.clearCanvas = function () {
            ctx.fillStyle = lineChart.backgroundColor;
            ctx.fillRect(0, 0, ctx.canvas.width, ctx.canvas.height);
        };

        lineChart.draw = function () {
            lineChart.clearCanvas();
            lineChart.drawAxes();
            lineChart.drawValues();
            lineChart.drawLegend();
            lineChart.drawSelectedValues();
        };

        lineChart.update = function () {
            if (lineChart.lines) {
                for (let [_lineName, line] of lineChart.lines) {
                    let overrun = line.values.length - lineChart.bufferSize;
                    if (overrun > 0) {
                        line.values = line.values.slice(overrun);
                    }
                }
            }
        };

        lineChart.tick = function (_timestamp) {
            if (lineChart.isDirty) {
                lineChart.update();
                lineChart.draw();
                lineChart.isDirty = false;
            }
        };

        ctx.canvas.addEventListener('click', function (evt) {
            lineChart.resetScale();
        });

        ctx.canvas.addEventListener('mouseenter', function (evt) {
            lineChart.shouldDrawSelectedValues = true;
        });

        ctx.canvas.addEventListener('mouseleave', function (evt) {
            lineChart.shouldDrawSelectedValues = false;
        });

        ctx.canvas.addEventListener('mousemove', function (evt) {
            lineChart.mousePosX = evt.clientX - lineChart.boundingRect.left;
            lineChart.mousePosY = evt.clientY - lineChart.boundingRect.top;
            lineChart.isDirty = true;
        });

        lineChart.clearCanvas();

        return lineChart;
    },
    initPacketInspector(id) {
        let ctl = document.getElementById(id);

        let initialState = JSON.parse(atob(ctl.dataset.initialState));

        let inspector = {};
        inspector.id = id;
        inspector.text = initialState.text ?? '';
        inspector.canvas = ctl.firstElementChild;
        inspector.context = inspector.canvas.getContext('2d');
        inspector.primaryColor = ctl.dataset.primaryColor ?? 'white';
        inspector.secondaryColor = ctl.dataset.secondaryColor ?? '#9e9baa';
        inspector.charHeight = initialState.char_height > 0 ? initialState.char_height : 28;
        inspector.byteCount = initialState.byte_count > 0 ? initialState.byte_count : 1;
        inspector.bitBoxes = [];
        inspector.selBitCount = 0;
        inspector.isDirty = false;

        let ctx = inspector.context;

        ctx.fillStyle = inspector.secondaryColor;
        ctx.fillRect(0, 0, inspector.canvas.width, inspector.canvas.height);
        ctx.fillStyle = inspector.primaryColor;
        ctx.strokeStyle = inspector.primaryColor;
        ctx.font = inspector.charHeight + 'px Courier';

        let charWidth = ctx.measureText('0').width;
        let row1y = inspector.charHeight;
        let row2y = inspector.charHeight * 2;
        let row3y = inspector.charHeight * 3;
        let row4y = inspector.charHeight * 4;
        let row5y = inspector.charHeight * 5;
        let row6y = inspector.charHeight * 6;
        let row5SelY = inspector.charHeight * 4.15;
        let row6SelY = inspector.charHeight * 5.15;
        let row7y = inspector.charHeight * 7;
        let row8y = inspector.charHeight * 8;
        let row9y = inspector.charHeight * 9;
        let columnWidth = 9 * charWidth;
        let columnStart = 5 * charWidth;
        let selResultColumnStart = 7 * charWidth;
        let totalWidth = columnStart + (inspector.byteCount * columnWidth);
        let selWidth = charWidth;
        let selHeight = inspector.charHeight;

        let canvasWidth = Number(inspector.canvas.getAttribute('width'));
        if (canvasWidth < totalWidth) {
            ctx.canvas.width = totalWidth;
            ctx.font = inspector.charHeight + 'px Courier';
        }

        ctx.fillStyle = inspector.secondaryColor;
        ctx.fillRect(0, 0, ctx.canvas.width, ctx.canvas.height);
        ctx.fillStyle = inspector.primaryColor;

        ctx.fillText('IDX', 5, row1y);
        ctx.fillText('HEX', 5, row2y);
        ctx.fillText('DEC', 5, row3y);
        ctx.fillText('BIN', 5, row4y);
        ctx.fillText('SEL', 5, row5y);
        ctx.fillText('INT', 5, row8y);
        ctx.fillText('UINT', 5, row9y);

        for (let i = 0; i < inspector.byteCount; i++) {
            let x = columnStart + (i * columnWidth);
            ctx.fillText(i.toString(), x, row1y);
        }

        let UInt8 = function (value) {
            return ((value >>> 0) & 0xFF);
        };

        let Int8 = function (value) {
            let ref = UInt8(value);
            return (ref > 0x7F) ? ref - 0x100 : ref;
        };

        let UInt16 = function (value) {
            return ((value >>> 0) & 0xFFFF);
        };

        let Int16 = function (value) {
            let ref = UInt16(value);
            return (ref > 0x7FFF) ? ref - 0x10000 : ref;
        };

        let UInt32 = function (value) {
            return ((value >>> 0) & 0xFFFFFFFF);
        };

        let Int32 = function (value) {
            let ref = UInt32(value);
            return (ref > 0x7FFFFFFF) ? ref - 0x100000000 : ref;
        };

        inspector.base64ToArray = function (base64) {
            let str = window.atob(base64);
            let len = str.length;
            let bytes = new Uint8Array(len);
            for (let i = 0; i < len; i++) {
                bytes[i] = str.charCodeAt(i);
            }
            return bytes;
        };

        inspector.reset = function () {

        };

        inspector.updateSelection = function () {
            let checkedCount = 0;
            for (let i = 0; i < inspector.bitBoxes.length; i++) {
                let bb = inspector.bitBoxes[i];
                if (!bb.checked) {
                    continue;
                }
                checkedCount++;

                let mask = 1 << (7 - (i % 8));
                let byte = inspector.bytes[(i / 8) >> 0];
                bb.binaryValue = (byte & mask) > 0 ? '1' : '0';
            }

            if (checkedCount > 32) {
                inspector.selectedValueAsInt = undefined;
                inspector.selectedValueAsUInt = undefined;
                return;
            }

            let selectedValueBits = Array(checkedCount);
            for (let bb of inspector.bitBoxes) {
                if (!bb.checked) {
                    continue;
                }
                selectedValueBits[bb.selectedIndex - 1] = bb.binaryValue;
            }

            let selectedValueBitString = '';
            for (let i = selectedValueBits.length - 1; i >= 0; i--) {
                selectedValueBitString += selectedValueBits[i];
            }

            if (checkedCount > 16) {
                let val = parseInt(selectedValueBitString.padStart(32, '0'), 2);
                inspector.selectedValueAsInt = Int32(val);
                inspector.selectedValueAsUInt = UInt32(val);
            } else if (checkedCount > 8) {
                let val = parseInt(selectedValueBitString.padStart(16, '0'), 2);
                inspector.selectedValueAsInt = Int16(val);
                inspector.selectedValueAsUInt = UInt16(val);
            } else {
                let val = parseInt(selectedValueBitString.padStart(8, '0'), 2);
                inspector.selectedValueAsInt = Int8(val);
                inspector.selectedValueAsUInt = UInt8(val);
            }

            GASP.sendEvent(GASP.newEvent(id, 'selection_update', {
                'int': inspector.selectedValueAsInt,
                'uint': inspector.selectedValueAsUInt
            }));

            return inspector;
        };

        inspector.updateBytes = function (packetBase64) {
            inspector.bytes = inspector.base64ToArray(packetBase64);
            inspector.isDirty = true;

            inspector.updateSelection();
            inspector.drawSelectionResult();

            return inspector;
        };

        inspector.drawBytes = function () {
            ctx.fillStyle = inspector.secondaryColor;
            ctx.fillRect(columnStart, 0, ctx.canvas.width, row4y);
            ctx.fillStyle = inspector.primaryColor;

            for (let i = 0; i < inspector.byteCount; i++) {
                let x = columnStart + (i * columnWidth);

                ctx.fillText(i.toString(), x, row1y);
                ctx.fillText(inspector.bytes[i].toString(16).padStart(2, '0'), x, row2y);
                ctx.fillText(inspector.bytes[i].toString(10).padStart(3, '0'), x, row3y);
                ctx.fillText(inspector.bytes[i].toString(2).padStart(8, '0'), x, row4y);
            }
        };

        inspector.drawSelectionResult = function () {
            ctx.fillStyle = inspector.secondaryColor;
            ctx.fillRect(selResultColumnStart, row7y, ctx.canvas.width, ctx.canvas.height);
            ctx.fillStyle = inspector.primaryColor;
            ctx.fillText(inspector.selectedValueAsInt.toString(), selResultColumnStart, row8y);
            ctx.fillText(inspector.selectedValueAsUInt.toString(), selResultColumnStart, row9y);
        };

        inspector.drawSelectionRow = function () {
            ctx.fillStyle = inspector.primaryColor;
            ctx.strokeStyle = inspector.primaryColor;

            let bitIndex = 0;
            for (let i = 0; i < inspector.byteCount; i++) {
                let x = columnStart + (i * columnWidth);

                for (let j = 0; j < 8; j++) {
                    let _x = x + (selWidth * j);

                    let bitBox = {};
                    bitBox.selectedIndex = 0;
                    bitBox.checked = false;
                    bitBox.x = _x;
                    bitBox.y = row5SelY;
                    bitBox.width = selWidth;
                    bitBox.height = selHeight;
                    inspector.bitBoxes.push(bitBox);

                    ctx.strokeRect(_x, row5SelY, selWidth, selHeight);

                    let bitIdx = bitIndex;
                    ctx.canvas.addEventListener('click', function (evt) {
                        if (evt.offsetX > _x && evt.offsetX < (_x + selWidth) && evt.offsetY > row5SelY && evt.offsetY < (row5SelY + inspector.charHeight)) {
                            let bb = inspector.bitBoxes[bitIdx];
                            let selIdx = bb.selectedIndex;
                            if (bb.checked) {
                                for (let k = 0; k < inspector.bitBoxes.length; k++) {
                                    let b = inspector.bitBoxes[k];
                                    if (b.checked && b.selectedIndex >= selIdx) {
                                        inspector.selBitCount--;
                                        b.checked = false;
                                        b.selectedIndex = 0;
                                        ctx.fillStyle = inspector.secondaryColor;
                                        ctx.fillRect(b.x + 1, row5SelY + 1, b.width - 2, b.height - 2);
                                        ctx.fillRect(b.x + 1, row6SelY + 1, b.width - 2, b.height - 2);
                                        ctx.fillStyle = inspector.primaryColor;
                                    }
                                }
                            } else {
                                inspector.selBitCount++;
                                bb.checked = true;
                                bb.selectedIndex = inspector.selBitCount;
                                let selBitCountStr = inspector.selBitCount.toString();
                                if (selBitCountStr.length > 1) {
                                    ctx.fillText(selBitCountStr[0], _x, row5y);
                                    ctx.fillText(selBitCountStr[1], _x, row6y);
                                } else {
                                    ctx.fillText(inspector.selBitCount.toString(), _x, row5y);
                                }
                            }

                            inspector.updateSelection();
                            inspector.drawSelectionResult();
                        }
                    });

                    bitIndex++;
                }
            }
        };

        inspector.draw = function (_timestamp) {
            if (inspector.isDirty) {
                inspector.drawBytes();
                inspector.isDirty = false;
            }
        };

        inspector.tick = function (_timestamp) {
            if (inspector.isDirty) {
                inspector.draw();
            }
        };

        inspector.drawSelectionRow();

        return inspector;
    },
    initComm(serverSocket, useTls) {
        if (useTls === undefined) {
            useTls = false;
        }
        /*tls_override*/

        let wsEndpoint = '';
        if (useTls) {
            wsEndpoint = 'wss://' + serverSocket + '/gaspws';
        } else {
            wsEndpoint = 'ws://' + serverSocket + '/gaspws';
        }

        this.socket = new WebSocket(wsEndpoint);

        this.socket.onopen = () => {
            console.log('Gasp: connected to server');
        };

        this.socket.onclose = () => {
            console.log('Gasp: disconnected from server');
        };

        this.socket.onmessage = msg => {
            let evt = JSON.parse(msg.data);
            switch (evt.type) {
                case 'form_update':
                    this.updateFormState(evt.data.state);
                    break;
                case 'textbox_update':
                    this.updateTextbox(evt.data);
                    break;
                case 'button_update':
                    this.updateButton(evt.data);
                    break;
                case 'label_update':
                    this.updateLabel(evt.data);
                    break;
                case 'dropdown_update':
                    this.updateDropdown(evt.data);
                    break;
                case 'checkbox_update':
                    this.updateCheckbox(evt.data);
                    break;
                case 'linechart_update':
                    this.updateLineChart(evt.data);
                    break;
                case 'packetinspector_update':
                    this.updatePacketInspector(evt.data);
                    break;
                default:
                    if (this.serverEventHandlers) {
                        if (this.serverEventHandlers[evt.type]) {
                            for (let i = 0; i < this.serverEventHandlers[evt.type].length; i++) {
                                this.serverEventHandlers[evt.type][i](evt);
                            }
                        }
                    }
                    break;
            }
        };

        this.socket.onerror = err => {
            console.log('Gasp: error: ' + JSON.stringify(err));
        };
    },
    init(uri) {
        if (this.initiated) {
            return;
        }

        this.initControls();
        this.initComm(uri);

        this.initiated = true;
    },
}

let GASP = Object.create(GASP_proto);
