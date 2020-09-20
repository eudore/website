package framework

import (
	"bufio"
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eudore/eudore"
	// "github.com/kr/pretty"
)

const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

type (
	Metas struct {
		ctx    eudore.Context
		Hooks  []StaticHook
		Labels []HeadLabel
	}
	StaticHook func(eudore.Context, HeadLabel) error
	HeadLabel  interface {
		Type() string
		String() string
	}
	titleLabel struct {
		title string
	}
	mapLabel struct {
		name   string
		data   string
		keys   map[string]string
		before bool
	}
	MergeFile struct {
		length      int64
		ext         string
		modtime     time.Time
		files       []*os.File
		multiReader io.Reader
	}
)

var (
	sha512Pool = sync.Pool{
		New: func() interface{} {
			return sha512.New()
		},
	}
)

func NewMetas(ctx eudore.Context, hooks []StaticHook) *Metas {
	meta := &Metas{
		ctx:   ctx,
		Hooks: hooks,
	}
	if lang := ctx.GetHeader("Accept-Language"); lang != "" {
		meta.AddMeta("language", lang)
	}
	if id := ctx.GetHeader("X-parent-Id"); id != "" {
		meta.AddMeta("parent-id", id)
	}
	if id := ctx.GetHeader("X-Request-Id"); id != "" {
		meta.AddMeta("request-id", id)
	}
	return meta
}

func (meta *Metas) AddLable(lab HeadLabel) {
	if lab == nil {
		return
	}
	for _, hook := range meta.Hooks {
		err := hook(meta.ctx, lab)
		if err != nil {
			meta.ctx.Errorf("Metas AddLable hook error: %v", err)
		}
	}
	meta.Labels = append(meta.Labels, lab)
}

func (meta *Metas) AddMeta(name, content string) {
	meta.Labels = append(meta.Labels, &mapLabel{
		data: fmt.Sprintf("<meta name='%s' content='%s'>", name, content),
	})
}

func NewLabel(str string) HeadLabel {
	if str == "" {
		return nil
	}
	var name = str
	{
		pos := strings.IndexByte(str, '>')
		posspace := strings.IndexByte(str, ' ')
		if posspace < pos {
			pos = posspace
		}
		if pos != -1 {
			name = str[1:pos]
		}
	}

	var lab HeadLabel
	switch name {
	case "title":
		lab = newTitleLable(str)
	default:
		lab = newMapLabel(str)
	}
	return lab
}

func newTitleLable(str string) HeadLabel {
	return &titleLabel{title: str[7 : len(str)-8]}
}

func (lab *titleLabel) Type() string {
	return "title"
}

func (lab *titleLabel) String() string {
	return fmt.Sprintf("<title>%s</title>", lab.title)
}

func newMapLabel(str string) HeadLabel {

	var lab mapLabel
	lab.keys = make(map[string]string)
	strs := strings.Split(str[1:len(str)-1], " ")
	lab.name = strs[0]

	laststr := strs[len(strs)-1]
	if strings.HasSuffix(laststr, fmt.Sprintf("</%s", strs[0])) {
		strs[len(strs)-1] = laststr[:len(laststr)-3-len(lab.name)]
		lab.before = true
	}
	for _, str := range strs[1:] {
		pos := strings.IndexByte(str, '=')
		if pos != -1 {

			lab.keys[str[:pos]] = str[pos+1:]
		} else {
			lab.keys[str] = ""
		}
	}
	return &lab
}

func (lab *mapLabel) Type() string {
	return lab.name
}

func (lab *mapLabel) String() string {
	if lab.name == "" {
		return lab.data
	}
	s := new(strings.Builder)
	s.WriteString("<" + lab.name)
	for key, val := range lab.keys {
		if val != "" {
			fmt.Fprintf(s, " %s=%s", key, val)
		} else {
			fmt.Fprintf(s, " %s", key)
		}
	}
	s.WriteByte('>')
	if lab.before {
		s.WriteString("</" + lab.name + ">")
	}
	return s.String()
}

func (lab mapLabel) Get(key string) string {
	return lab.keys[key]
}

func WithPushHook(ctx eudore.Context, lab HeadLabel) error {

	var path string
	switch lab.Type() {
	case "script":
		path = lab.(*mapLabel).Get("src")
	case "link":
		path = lab.(*mapLabel).Get("herf")
	}
	if path != "" {
		ctx.Push(path[1:len(path)-1], nil)
	}
	return nil

}

func WithSRIHook(ctx eudore.Context, lab HeadLabel) error {
	maplab, _ := lab.(*mapLabel)
	switch lab.Type() {
	case "script":
		sri, err := GetSha512Value(maplab.keys["src"])
		if err != nil {
			return err
		}

		maplab.keys["integrity"] = sri
	case "link":
		sri, err := GetSha512Value(maplab.keys["href"])
		if err != nil {
			return err
		}
		maplab.keys["integrity"] = sri
	}
	return nil

}

// 计算SRI512
func GetSha512Value(path string) (string, error) {
	// read
	read, err := GetPathBody("static" + path[1:len(path)-1])
	if err != nil {
		return "", err
	}

	h := sha512Pool.Get().(hash.Hash)
	h.Reset()
	_, err = io.Copy(h, read)

	var val string
	if err == nil {
		val = "sha512-" + base64.StdEncoding.EncodeToString(h.Sum(nil))
	}
	sha512Pool.Put(h)
	// read.Close()
	return val, err
}

func GetPathBody(path string) (io.ReadCloser, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		response, err := http.Get(path)
		return response.Body, err
	}
	// return os.Open(path)
	mf, err := OpenMergeFile(path)
	return mf, err
}

func checkIfModifiedSince(ctx eudore.Context, modtime time.Time) bool {
	if ctx.Method() != "GET" && ctx.Method() != "HEAD" {
		return false
	}
	ims := ctx.GetHeader("If-Modified-Since")
	if ims == "" || isZeroTime(modtime) {
		return false
	}
	t, err := http.ParseTime(ims)
	if err != nil {
		return false
	}

	// The Date-Modified header truncates sub-second precision, so
	// use mtime < t+1s instead of mtime <= t to check for unmodified.
	if modtime.Before(t.Add(1 * time.Second)) {
		return false
	}
	return true
}

var unixEpochTime = time.Unix(0, 0)

// isZeroTime reports whether t is obviously unspecified (either zero or Unix()=0).
func isZeroTime(t time.Time) bool {
	return t.IsZero() || t.Equal(unixEpochTime)
}

func OpenMergeFile(path string) (*MergeFile, error) {
	fs := &MergeFile{}

	pos := strings.LastIndexByte(path, '/') + 1
	dir, name := path[:pos], path[pos:]

	pos = strings.IndexByte(name, '.')
	ext := ""
	if pos != -1 {
		posl := strings.LastIndexByte(name, '.')
		fs.ext = mime.TypeByExtension(name[posl:])
		ext = name[pos:]
		name = name[:pos]
	}
	paths := strings.Split(name, "-")
	for i := range paths {
		paths[i] = dir + paths[i] + ext
	}

	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			fs.Close()
			return nil, err
		}

		desc, err := f.Stat()
		if err != nil {
			fs.Close()
			return nil, err
		}

		if desc.ModTime().Sub(fs.modtime) > 0 {
			fs.modtime = desc.ModTime()
		}

		fs.files = append(fs.files, f)
		fs.length += desc.Size()
	}

	return fs, nil
}

func (mf *MergeFile) WriteTo(w io.Writer) (int64, error) {
	var length int64
	for _, file := range mf.files {
		n, err := io.Copy(w, file)
		if err != nil && err != io.EOF {
			return length, err
		}
		w.Write([]byte{'\r', '\n'})
		length = length + n + 2
	}
	return length, nil
}

func (mf *MergeFile) Read(p []byte) (n int, err error) {
	for len(mf.files) > 0 {
		n, err = mf.files[0].Read(p)
		if err == io.EOF {
			err = mf.files[0].Close()
			// Use eofReader instead of nil to avoid nil panic
			// after performing flatten (Issue 18232).
			mf.files = mf.files[1:]
		}
		if n > 0 || err != io.EOF {
			if err == io.EOF && len(mf.files) > 0 {
				// Don't return EOF yet. More readers remain.
				err = nil
			}
			return
		}
	}
	return 0, io.EOF
}

func (mf *MergeFile) Close() error {
	errs := NewErrors()
	for _, r := range mf.files {
		errs.HandleError(r.Close())
	}
	return errs.GetError()
}

func AutoInjectHTML(router eudore.Router, dir string) error {
	names, err := readAllFiles("", dir)
	if err != nil {
		return err
	}
	for _, name := range names {
		if strings.HasSuffix(name, "index.html") {
			name := strings.TrimSuffix(name, "index.html")
			route := router.Params().Get("route") + name
			router.GetFunc(name, NewHTMLHandlerFunc(path.Join(dir, name, "index.html")))
			router.GetFunc(name+"*path", func(ctx eudore.Context) {
				ctx.Redirect(302, fmt.Sprintf("%s#/%s", route, ctx.Params().Get("path")))
			})
		} else {
			router.GetFunc(strings.TrimSuffix(name, ".html"), NewHTMLHandlerFunc(path.Join(dir, name)))
		}
	}
	return nil
}

func readAllFiles(perfix, dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(files))
	for _, file := range files {
		name := file.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if file.IsDir() {
			newnames, err := readAllFiles(perfix+"/"+name, dir+"/"+name)
			if err != nil {
				return nil, err
			}
			names = append(names, newnames...)
		} else {
			if strings.HasSuffix(name, ".html") {
				names = append(names, perfix+"/"+name)
			}
		}
	}

	return names, nil
}

func NewHTMLHandlerFunc(path string) func(eudore.Context) error {
	hooks := []StaticHook{WithPushHook, WithSRIHook}
	return func(ctx eudore.Context) error {
		meta := NewMetas(ctx, hooks)

		ctx.SetHeader("Content-Type", "text/html; charset=utf-8")

		// open file
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		desc, err := f.Stat()
		if err != nil {
			return err
		}

		// add header
		h := ctx.Response().Header()
		h.Set("Last-Modified", desc.ModTime().UTC().Format(TimeFormat))

		hr := bufio.NewReader(f)
		hw := bytes.NewBuffer(nil)
		var inhead bool
		var space []byte
		for {
			str, err := hr.ReadString('\n')
			if err != nil {
				break
			}
			if strings.TrimSpace(str) == "<head>" {
				space = []byte(str[:strings.IndexByte(str, '<')])
				inhead = true
				hw.WriteString(str)
			} else if strings.TrimSpace(str) == "</head>" {
				for _, meta := range meta.Labels {
					fmt.Fprintf(hw, "%s\t%s\n", space, meta.String())
				}
				fmt.Fprintf(hw, "%s</head>\n", space)
				break
			} else if inhead {
				lab := NewLabel(strings.TrimSpace(str))
				meta.AddLable(lab)
			} else {
				hw.WriteString(str)
			}
		}

		_, err = io.Copy(ctx, hw)

		io.Copy(ctx, hr)
		return err
	}
}

func NewMergeFileHandlerFunc(perfix string) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		upath := ctx.GetParam("path")
		if upath == "" {
			upath = ctx.Path()
		}

		mf, err := OpenMergeFile(path.Join(perfix, path.Clean("/"+upath)))
		if err != nil {
			ctx.Fatal(err)
			return
		}

		defer mf.Close()
		// check cache
		if checkIfModifiedSince(ctx, mf.modtime) {
			ctx.WriteHeader(eudore.StatusNotModified)
			return
		}

		h := ctx.Response().Header()
		h.Set("Last-Modified", mf.modtime.UTC().Format(TimeFormat))
		h.Set("Content-Type", mf.ext)
		if h.Get("Content-Encoding") == "" {
			h.Set("Content-Length", strconv.FormatInt(mf.length+int64(len(mf.files)), 10))
		}

		_, err = io.Copy(ctx, mf)
		if err != nil {
			ctx.Fatal((err))
		}
	}
}

func NewStaticHandlerFunc(perfix string) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		upath := ctx.GetParam("path")
		if upath == "" {
			upath = ctx.Path()
		}
		path := path.Join(perfix, path.Clean("/"+upath))

		files, err := ioutil.ReadDir(path)
		if err == nil {
			ctx.SetHeader("Content-Type", "text/html; charset=utf-8")
			ctx.WriteString(fmt.Sprintf(`<html>
<head><title>Index of %s</title></head>
<body bgcolor='white'>
<h1>Index of %s</h1><hr><pre><a href='../'>../</a>`, ctx.Path(), ctx.Path()))

			for _, file := range files {
				name, times, size, isdir := file.Name(), file.ModTime(), strconv.FormatInt(file.Size(), 10), file.IsDir()
				if isdir {
					name += "/"
					size = "-"
				}
				ctx.WriteString(fmt.Sprintf("\n<a href='%s'>%s</a>%s%s       %s", name, name, strings.Repeat(" ", 50-len(name)), times.Format("2-Jan-2006 15:04:05"), size))
			}

			ctx.WriteString("\n</pre><hr></body>\n</html>")
		} else if err.Error() == "readdirent: not a directory" {
			ctx.WriteFile(path)
		} else {
			ctx.Fatal(err)
		}
	}
}
