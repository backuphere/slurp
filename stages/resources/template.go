package resources

import (
	"bytes"
	"fmt"
	"io"
	"text/template"
)

var pkg *template.Template

func reader(input io.Reader) (string, error) {

	var (
		buff       bytes.Buffer
		err        error
		blockwidth int = 12
		curblock   int = 0
	)

	b := make([]byte, blockwidth)

	for n, err := input.Read(b); err == nil; n, err = input.Read(b) {
		for i := 0; i < n; i++ {
			fmt.Fprintf(&buff, "0x%02x,", b[i])
			curblock++
			if curblock < blockwidth {
				continue
			}
			buff.Write([]byte{'\n'})
			buff.Write([]byte{'\t', '\t'})
			curblock = 0
		}
	}

	return buff.String(), err
}

func init() {

	pkg = template.Must(template.New("file").Funcs(template.FuncMap{"reader": reader}).Parse(` &File{
  Reader: bytes.NewReader([]byte{ {{ reader . }} }),
  name:    "{{ .Stat.Name }}", 
  size:    {{ .Stat.Size }},
  modTime: time.Unix({{ .Stat.ModTime.Unix }},{{ .Stat.ModTime.UnixNano }}),
  isDir:   {{ .Stat.IsDir }},
}`))

	pkg = template.Must(pkg.New("pkg").Parse(`//Generated by slurp/resources
package {{ .Pkg }}

import (
  "errors"
  "net/http"
  "time"
  "bytes"
  "os"
)


var (
    {{ .Var }} *FileSystem
	ErrNotFound = errors.New("File Not Found.")
)

// Helper functions for easier file access.
func Open(name string) (http.File, error) {
	return {{ .Var }}.Open(name)
}

// http.FileSystem implementation.
type FileSystem struct {
	files map[string]*File
}

func (fs *FileSystem) Open(name string) (http.File, error) {
	var err error
	file, ok := fs.files[name]
	if !ok {
		err = ErrNotFound
	}
	return file, err
}

type File struct {
	*bytes.Reader
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

// A noop-closer.
func (f *File) Close() error {
	return nil
}

func (f *File) Readdir(count int) ([]os.FileInfo, error) {
  return nil, errors.New("Not Supported.")
}


func (f *File) Stat() (os.FileInfo, error) {
	return &FileInfo{
		name:    f.name,
		size:    f.size,
		mode:    f.mode,
		modTime: f.modTime,
		isDir:   f.isDir,
		sys:     f.Reader,
	}, nil
}

type FileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
	sys     interface{}
}

func (f *FileInfo) Name() string {
	return f.name
}
func (f *FileInfo) Size() int64 {
	return f.size
}

func (f *FileInfo) Mode() os.FileMode {
	return f.mode
}

func (f *FileInfo) ModTime() time.Time {
	return f.modTime
}

func (f *FileInfo) IsDir() bool {
	return f.isDir
}

func (f *FileInfo) Sys() interface{} {
	return f.sys
}


func init() {
  {{ .Var }} = &FileSystem{
		files: map[string]*File{
		  {{range .Files }} "{{.Stat.Name}}": {{ template "file" . }}, {{ end }}
		},
	  }
}
`))
}
