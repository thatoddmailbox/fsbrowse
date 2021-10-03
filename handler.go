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
	d, ok := dir.(fs.ReadDirFile)
	if !ok {
		w.Write([]byte("file does not readdir"))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	direntries, err := d.ReadDir(0)
	if err != nil {
		panic(err)
	}

	h.dirTemplate.Execute(w, direntries)
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	dir, err := h.root.Open(path)
	if err != nil {
		if err == fs.ErrNotExist {
			w.WriteHeader(404)
			w.Write([]byte("file not found"))
			return
		}

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
