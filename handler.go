package fsbrowse

import (
	_ "embed"
	"errors"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

//go:embed dir.html
var dirTemplateSource string
var dirTemplate *template.Template

type handler struct {
	root fs.FS

	prefix string
	notice string
}

type Options struct {
	Prefix string
	Notice string
}

func serveDir(dir fs.File, s fs.FileInfo, pathParts []string, w http.ResponseWriter, prefix string, notice string) {
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
	err = dirTemplate.Execute(w, map[string]interface{}{
		"dir":        s,
		"direntries": direntries,
		"pathParts":  pathParts,
		"notice":     notice,
		"prefix":     prefix,
	})
	if err != nil {
		panic(err)
	}
}

func doServeHTTP(w http.ResponseWriter, r *http.Request, root fs.FS, prefix string, notice string) {
	path := r.URL.Path
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	pathParts := strings.Split(path, "/")

	if len(path) == 0 {
		path = "."
	}

	f, err := root.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
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
		serveDir(f, s, pathParts, w, prefix, notice)
		return
	}

	rs, ok := f.(io.ReadSeeker)
	if !ok {
		w.Write([]byte("fs file does not support Seek"))
		return
	}

	http.ServeContent(w, r, s.Name(), s.ModTime(), rs)
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	doServeHTTP(w, r, h.root, h.prefix, h.notice)
}

func ServeHTTPStateless(w http.ResponseWriter, r *http.Request, root fs.FS, prefix string, notice string) {
	initTemplate()
	doServeHTTP(w, r, root, prefix, notice)
}

func initTemplate() {
	if dirTemplate != nil {
		return
	}

	var err error
	dirTemplate, err = template.New("dir").Funcs(template.FuncMap{
		"formatSize": func(s int64) string {
			prefixes := []string{"", "K", "M", "G", "T"}
			prefix := 0
			for s > 1000 && prefix < len(prefixes)-1 {
				s /= 1000
				prefix++
			}

			return strconv.FormatInt(s, 10) + " " + prefixes[prefix] + "B"
		},
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05 -0700 MST")
		},
	}).Parse(dirTemplateSource)
	if err != nil {
		panic(err)
	}
}

func FileServerWithOptions(root fs.FS, options Options) http.Handler {
	initTemplate()

	return &handler{
		root:   root,
		prefix: options.Prefix,
		notice: options.Notice,
	}
}

func FileServer(root fs.FS) http.Handler {
	return FileServerWithOptions(root, Options{
		Prefix: "",
		Notice: "",
	})
}
