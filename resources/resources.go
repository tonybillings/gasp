package resources

import _ "embed"

//go:embed html/form.html
var FormTemplate string

//go:embed css/gasp.css
var GaspStyle string

//go:embed js/gasp.js
var GaspScript string
