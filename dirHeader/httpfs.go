package dirHeader

import (
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"

	"github.com/inhies/go-bytesize"
)

type Data struct {
	Header DirHeader
	Rows   []Row
}

func ReadDirectory(fullPath, uPath string) (*Data, error) {
	dirs, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name() < dirs[j].Name() })

	header := DirHeader{
		TextDirection: "ltr",
		Language:      "en",

		Header: "Index of " + uPath,

		HeaderName:         "Name",
		HeaderSize:         "Size",
		HeaderDateModified: "Date Modified",
	}
	if uPath != "/" {
		header.ParentDirText = "[parent directory]"
	}
	rows := make([]Row, 0)
	for _, file := range dirs {
		name := file.Name()
		var size int64
		var modTime time.Time
		var isDir bool
		info, err := file.Info()
		if err != nil {
			size = 0
			modTime = time.Unix(0, 0)
			isDir = false
		} else {
			size = info.Size()
			modTime = info.ModTime()
			if info.IsDir() {
				isDir = true
			} else {
				isDir = false
				if (info.Mode() & fs.ModeSymlink) == fs.ModeSymlink {
					link, err := filepath.EvalSymlinks(path.Join(fullPath, name))
					if err == nil {
						info, err = os.Stat(link)
						if err == nil {
							size = info.Size()
							modTime = info.ModTime()
							isDir = info.IsDir()
						}
					}
				}
			}
		}
		u := url.URL{Path: name}
		rows = append(rows, Row{
			Name:               name,
			Url:                u.String(),
			IsDir:              isDir,
			Size:               size,
			SizeString:         bytesize.ByteSize(size).String(),
			DateModified:       modTime.Unix(),
			DateModifiedString: modTime.Format("2006/01/02 15:04:05"),
		})
	}

	return &Data{
		Header: header,
		Rows:   rows,
	}, nil
}

func DirList(w http.ResponseWriter, fullPath, uPath string) (int, string) {
	data, err := ReadDirectory(fullPath, uPath)
	if err != nil {
		http.Error(w, "Error reading directory", http.StatusInternalServerError)
		return http.StatusInternalServerError, "error reading directory: " + err.Error()
	}

	html, err := DirHtml(&data.Header, data.Rows)
	if err != nil {
		http.Error(w, "Error render", http.StatusInternalServerError)
		return http.StatusInternalServerError, "error render: " + err.Error()
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = w.Write(html)
	if err != nil {
		http.Error(w, "Error writing body", http.StatusInternalServerError)
		return http.StatusInternalServerError, "error writing body: " + err.Error()
	}
	return http.StatusOK, ""
}
