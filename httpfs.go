package main

import (
	"io/fs"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/inhies/go-bytesize"
)

type anyDirs interface {
	len() int
	name(i int) string
	isDir(i int) bool
	size(i int) int64
	modTime(i int) time.Time
}

type fileInfoDirs []fs.FileInfo

func (d fileInfoDirs) len() int                { return len(d) }
func (d fileInfoDirs) isDir(i int) bool        { return d[i].IsDir() }
func (d fileInfoDirs) name(i int) string       { return d[i].Name() }
func (d fileInfoDirs) size(i int) int64        { return d[i].Size() }
func (d fileInfoDirs) modTime(i int) time.Time { return d[i].ModTime() }

type dirEntryDirs []fs.DirEntry

func (d dirEntryDirs) len() int          { return len(d) }
func (d dirEntryDirs) isDir(i int) bool  { return d[i].IsDir() }
func (d dirEntryDirs) name(i int) string { return d[i].Name() }
func (d dirEntryDirs) size(i int) int64 {
	info, err := d[i].Info()
	if err != nil {
		return 0
	}
	return info.Size()
}
func (d dirEntryDirs) modTime(i int) time.Time {
	info, err := d[i].Info()
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

func DirList(w http.ResponseWriter, r *http.Request, f http.File, path string) (int, string) {
	// Prefer to use ReadDir instead of Readdir,
	// because the former doesn't require calling
	// Stat on every entry of a directory on Unix.

	var dirs anyDirs
	var err error
	if d, ok := f.(fs.ReadDirFile); ok {
		var list dirEntryDirs
		list, err = d.ReadDir(-1)
		dirs = list
	} else {
		var list fileInfoDirs
		list, err = f.Readdir(-1)
		dirs = list
	}

	if err != nil {
		http.Error(w, "Error reading directory", http.StatusInternalServerError)
		return http.StatusInternalServerError, "error reading directory: "+err.Error()
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs.name(i) < dirs.name(j) })

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	header := DirHeader{
		TextDirection: "ltr",
		Language:      "zh",

		Header: path + " 的索引",

		HeaderName:         "名称",
		HeaderSize:         "大小",
		HeaderDateModified: "修改日期",
	}
	if path != "/" {
		header.ParentDirText = "[上级目录]"
	}
	rows := make([]Row, 0)
	for i, n := 0, dirs.len(); i < n; i++ {
		name := dirs.name(i)
		u := url.URL{Path: name}
		size := dirs.size(i)
		t := dirs.modTime(i)
		rows = append(rows, Row{
			Name:               name,
			Url:                u.String(),
			IsDir:              dirs.isDir(i),
			Size:               dirs.size(i),
			SizeString:         bytesize.ByteSize(size).String(),
			DateModified:       t.Unix(),
			DateModifiedString: t.Format("2006/01/02 15:04:05"),
		})
	}
	html, err := DirHtml(&header, rows)
	if err != nil {
		http.Error(w, "Error render", http.StatusInternalServerError)
		return http.StatusInternalServerError, "error render: "+err.Error()
	}
	_, err = w.Write(html)
	if err != nil {
		http.Error(w, "Error writing body", http.StatusInternalServerError)
		return http.StatusInternalServerError, "error writing body: "+err.Error()
	}
	return http.StatusOK, ""
}
