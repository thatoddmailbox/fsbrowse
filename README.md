# fsbrowse
A web UI to browse an [fs.FS](https://pkg.go.dev/io/fs#FS). Effectively a slightly nicer-looking version of [http.FileServer](https://pkg.go.dev/net/http#FileServer).

Requires Go 1.16 or newer, due to the usage of fs.FS. Also, the fs.FS being used must implement io.Seeker on its files. For directory browsing to work, the fs.FS must also implement fs.ReadDirFile on its files.