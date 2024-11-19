package main

import (
	"encoding/json"
	"go-ssr/common"
	"html/template"
	"log"
	"os"
	tmpl "text/template"

	"github.com/dop251/goja"
	esbuild "github.com/evanw/esbuild/pkg/api"
)

// [Yaffle/TextEncoderTextDecoder.js](https://gist.github.com/Yaffle/5458286)
var textEncoderPolyfill = `function TextEncoder(){} TextEncoder.prototype.encode=function(string){var octets=[],length=string.length,i=0;while(i<length){var codePoint=string.codePointAt(i),c=0,bits=0;codePoint<=0x7F?(c=0,bits=0x00):codePoint<=0x7FF?(c=6,bits=0xC0):codePoint<=0xFFFF?(c=12,bits=0xE0):codePoint<=0x1FFFFF&&(c=18,bits=0xF0),octets.push(bits|(codePoint>>c)),c-=6;while(c>=0){octets.push(0x80|((codePoint>>c)&0x3F)),c-=6}i+=codePoint>=0x10000?2:1}return octets};function TextDecoder(){} TextDecoder.prototype.decode=function(octets){var string="",i=0;while(i<octets.length){var octet=octets[i],bytesNeeded=0,codePoint=0;octet<=0x7F?(bytesNeeded=0,codePoint=octet&0xFF):octet<=0xDF?(bytesNeeded=1,codePoint=octet&0x1F):octet<=0xEF?(bytesNeeded=2,codePoint=octet&0x0F):octet<=0xF4&&(bytesNeeded=3,codePoint=octet&0x07),octets.length-i-bytesNeeded>0?function(){for(var k=0;k<bytesNeeded;){octet=octets[i+k+1],codePoint=(codePoint<<6)|(octet&0x3F),k+=1}}():codePoint=0xFFFD,bytesNeeded=octets.length-i,string+=String.fromCodePoint(codePoint),i+=bytesNeeded+1}return string};`
var processPolyfill = `var process = {env: {NODE_ENV: "production"}};`
var consolePolyfill = `var console = {log: function(){}};`

type InitialProps struct {
	Name          string
	InitialNumber int
}

func buildBackend() string {
	result := esbuild.Build(esbuild.BuildOptions{
		EntryPoints:       []string{"frontend/serverEntry.jsx"},
		Bundle:            true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Write:             false,
		Outdir:            "/",
		Format:            esbuild.FormatIIFE,
		Platform:          esbuild.PlatformBrowser,
		Target:            esbuild.ES2015,
		Banner: map[string]string{
			"js": textEncoderPolyfill + processPolyfill + consolePolyfill,
		},
		Loader: map[string]esbuild.Loader{
			".jsx": esbuild.LoaderJSX,
		},
	})
	script := string(result.OutputFiles[0].Contents)
	return script
}

func buildClient() {
	clientResult := esbuild.Build(esbuild.BuildOptions{
		EntryPoints:       []string{"frontend/clientEntry.jsx"},
		Bundle:            true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Write:             true,
		Format:            esbuild.FormatESModule,
	})
	clientBundleString := string(clientResult.OutputFiles[0].Contents)
	err := os.MkdirAll("dist", 0755)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("dist/app.js")
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	f.WriteString(clientBundleString)
}

func gen() (*common.PageData, error) {
	// Create a bundle for SSR in react
	backendBundle := buildBackend()
	// Create a bundle for react hydrate
	buildClient()

	vm := goja.New()
	_, err := vm.RunScript("bundle.js", backendBundle)
	if err != nil {
		return nil, err
	}

	val, err := vm.RunScript("render.js", "renderApp()")
	if err != nil {
		return nil, err
	}

	renderedHTML := val.String()
	// Props passed from the backend
	initialProps := InitialProps{
		Name:          "SSR for React in Go",
		InitialNumber: 100,
	}

	jsonProps, err := json.Marshal(initialProps)
	if err != nil {
		return nil, err
	}

	data := common.PageData{
		RenderedContent: template.HTML(renderedHTML),
		InitialProps:    template.JS(jsonProps),
	}

	return &data, nil
}

func main() {
	data, err := gen()
	if err != nil {
		log.Fatal(err)
	}

	temp, err := os.ReadFile("code-gen/templ.go.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	result, err := os.Create("common/data.go")
	if err != nil {
		log.Fatal(err)
	}

	defer result.Close()

	t := tmpl.Must(tmpl.New("data").Parse(string(temp)))
	err = t.Execute(result, data)
	if err != nil {
		os.Remove("common/data.go")
		log.Fatal(err)
	}
}
