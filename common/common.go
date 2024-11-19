package common

import "html/template"

type PageData struct {
	RenderedContent template.HTML
	InitialProps    template.JS
	JS              template.JS
}
