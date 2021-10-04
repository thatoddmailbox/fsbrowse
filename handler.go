package fsbrowse

import (
	_ "embed"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

//go:embed dir.html
var dirTemplateSource string

type handler struct {
	root fs.FS

	dirTemplate *template.Template
}

func (h *handler) serveDir(dir fs.File, s fs.FileInfo, pathParts []string, w http.ResponseWriter) {
	d, ok := dir.(fs.ReadDirFile)
	if !ok {
		w.Write([]byte("file does not readdir"))
		return
	}

	direntries, err := d.ReadDir(0)
	if err != nil {
		panic(err)
	}

	sort.Slice(direntries, func(i, j int) bool {
		return direntries[i].Name() < direntries[j].Name()
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = h.dirTemplate.Execute(w, map[string]interface{}{
		"dir":        s,
		"direntries": direntries,
		"pathParts":  pathParts,
	})
	if err != nil {
		panic(err)
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	pathParts := strings.Split(path, "/")

	f, err := h.root.Open(path)
	if err != nil {
		if err == fs.ErrNotExist {
			w.WriteHeader(404)
			w.Write([]byte("file not found"))
			return
		}

		panic(err)
	}

	defer f.Close()

	s, err := f.Stat()
	if err != nil {
		panic(err)
	}

	if s.IsDir() {
		h.serveDir(f, s, pathParts, w)
		return
	}

	rs, ok := f.(io.ReadSeeker)
	if !ok {
		w.Write([]byte("fs file does not support Seek"))
		return
	}

	http.ServeContent(w, r, s.Name(), s.ModTime(), rs)
}

func FileServer(root fs.FS) http.Handler {
	dirTemplate, err := template.New("dir").Funcs(template.FuncMap{
		"formatSize": func(s int64) string {
			prefixes := []string{"", "K", "M", "G", "T"}
			prefix := 0
			for s > 1000 && prefix < len(prefixes)-1 {
				s /= 1000
				prefix++
			}

			return strconv.FormatInt(s, 10) + " " + prefixes[prefix] + "B"
		},
	}).Parse(dirTemplateSource)
	if err != nil {
		panic(err)
	}

	return &handler{
		root: root,

		dirTemplate: dirTemplate,
	}
}
