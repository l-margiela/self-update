package main

import (
	"html/template"
	"net/http"

	"go.uber.org/zap"
)

var (
	pageTemplate = `
<!DOCTYPE html>
<html>
<head>
<title> Server {{.Version}} </title> </head>
<body>
<h1>This server is version {{.Version}}</h1>
<a href="check">Check for new version</a>
<br>
{{if .NewVersion}}New version is available: {{.NewVersion}} | <a
href="upgrade">Upgrade</a>{{end}} </body>
</html>
`
)

type Status struct {
	Version    string
	NewVersion string
}

var compiledPage *template.Template

// compilePage returns compiled `pageTemplate`.
//
// It uses memoisation to cache the compilation result.
//
// The design here depends on global variables which is a bad pattern
// although given the project's size and scope, it is acceptable.
func compilePage() *template.Template {
	if compiledPage != nil {
		return compiledPage
	}

	var err error
	compiledPage, err = template.New("page").Parse(pageTemplate)
	if err != nil {
		// This is one of a few places where panic() is an idiomatic approach.
		// The pageTemplate does not change in runtime, so an invalid template
		// means that the program itself is invalid.
		panic(err)
	}
	return compiledPage
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	zap.L().Info("handle HTTP request", zap.String("method", r.Method), zap.String("uri", r.RequestURI))
	page := compilePage()
	if err := page.Execute(w, Status{Version, ""}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte(err.Error())); err != nil {
			zap.L().Error("write response", zap.Error(err))
		}

		zap.L().Error("handle /", zap.Error(err))
		return
	}
}
