package main

import (
	"fmt"
	"go-ssr/common"
	"html/template"
	"log"
	"net/http"
)

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>React App</title>
</head>
<body>
    <div id="app">{{.RenderedContent}}</div>
	<script type="module" src="static/app.js"></script>
	<script>window.APP_PROPS = {{.InitialProps}};</script>
</body>
</html>
`

type PageData struct {
	RenderedContent template.HTML
	InitialProps    template.JS
	JS              template.JS
}

type InitialProps struct {
	Name          string
	InitialNumber int
}

//go:generate go run code-gen/main.go
func main() {
	tmpl, err := template.New("webpage").Parse(htmlTemplate)
	if err != nil {
		log.Fatal("Error parsing template:", err)
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("dist/"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		err := tmpl.Execute(w, common.Data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	fmt.Println("Server is running at http://localhost:3002")
	log.Fatal(http.ListenAndServe(":3002", nil))
}
