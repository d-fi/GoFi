package web

import _ "embed"

//go:embed index.html
var indexHTML string

//go:embed style.css
var styleCSS string

//go:embed script.js
var scriptJS string
