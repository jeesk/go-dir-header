package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
)

type fileHandler struct {
	root http.FileSystem
}

func FileServer(root http.FileSystem) http.Handler {
	return &fileHandler{root}
}

func (f *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}
	uPath := r.URL.Path
	if !strings.HasPrefix(uPath, "/") {
		uPath = "/" + uPath
		r.URL.Path = uPath
	}
	uPath = path.Clean(uPath)

	file, err := f.root.Open(uPath)
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}

	if info.IsDir() {
		DirList(w, r, file, uPath)
		return
	}

	http.ServeContent(w, r, info.Name(), info.ModTime(), file)
}

// toHTTPError returns a non-specific HTTP error message and status code
// for a given non-nil error value. It's important that toHTTPError does not
// actually return err.Error(), since msg and httpStatus are returned to users,
// and historically Go's ServeContent always returned just "404 Not Found" for
// all errors. We don't want to start leaking information in error messages.
func toHTTPError(err error) (msg string, httpStatus int) {
	if os.IsNotExist(err) {
		return "404 page not found", http.StatusNotFound
	}
	if os.IsPermission(err) {
		return "403 Forbidden", http.StatusForbidden
	}
	// Default:
	return "500 Internal Server Error", http.StatusInternalServerError
}

func main() {
	bind := flag.String("bind", "0.0.0.0:8080", "listen address")
	root := flag.String("root", "/", "root directory")
	flag.Parse()

	addr := *bind
	http.Handle("/", FileServer(http.Dir(*root)))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		_, _= fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Serving HTTP http://%s -> %s\n", listener.Addr().String(), *root)
	err = http.Serve(listener, nil)
	if err != nil {
		_, _= fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
