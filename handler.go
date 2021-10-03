package fsbrowse

import (
	_ "embed"
	"html/template"
	"io/fs"
	"net/http"
)

//go:embed dir.html
var dirTemplateSource string

type handler struct {
	root fs.FS

	dirTemplate *template.Template
}

func (h *handler) serveDir(dir fs.File, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.dirTemplate.Execute(w, dir)
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dir, err := h.root.Open(".")
	if err != nil {
		panic(err)
	}

	defer dir.Close()
	h.serveDir(dir, w)
}

func FileServer(root fs.FS) http.Handler {
	dirTemplate, err := template.New("dir").Parse(dirTemplateSource)
	if err != nil {
		panic(err)
	}

	return &handler{
		root: root,

		dirTemplate: dirTemplate,
	}
}
