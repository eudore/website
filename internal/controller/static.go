package controller

/*
staic控制器执行静态文件相关内容。
*/
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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eudore/eudore"
	// "github.com/kr/pretty"
)

type (
	StaticHook       func(*ControllerStatic, HeadLabel)
	ControllerStatic struct {
		eudore.Context
		Hook     StaticHook
		hashPool sync.Pool
		Meta     []HeadLabel
	}
	HeadLabel interface {
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

func NewControllerStatic() *ControllerStatic {
	ctl := &ControllerStatic{
		Hook: func(*ControllerStatic, HeadLabel) {},
		hashPool: sync.Pool{
			New: func() interface{} {
				return sha512.New()
			},
		},
	}
	ctl.WithPushHook()
	ctl.WithSRIHook()
	return ctl
}

func (ctl *ControllerStatic) Init(ctx eudore.Context) error {
	ctl.Context = ctx
	ctl.Meta = ctl.Meta[0:0]

	if lang := ctx.GetHeader("Accept-Language"); lang != "" {
		ctl.AddMeta("language", lang)
	}
	if id := ctx.GetHeader("X-parent-Id"); id != "" {
		ctl.AddMeta("parent-id", id)
	}
	if id := ctx.GetHeader("X-Request-Id"); id != "" {
		ctl.AddMeta("request-id", id)
	}
	return nil
}
func (ctl *ControllerStatic) Clone() *ControllerStatic {
	return &ControllerStatic{
		Hook:     ctl.Hook,
		hashPool: ctl.hashPool,
	}
}

func (ctl *ControllerStatic) WriteHTML(path string) error {
	ctl.PushAll(path)
	ctl.SetHeader("Content-Type", "text/html; charset=utf-8")

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
	h := ctl.Response().Header()
	h.Set("Last-Modified", desc.ModTime().UTC().Format(eudore.TimeFormat))

	var br = bufio.NewReader(f)
	var b bytes.Buffer
	var inhead bool
	var space []byte
	for {
		str, err := br.ReadString('\n')
		if err != nil {
			ctl.Error(err)
			return err
		}
		if strings.TrimSpace(str) == "<head>" {
			space = []byte(str[:strings.IndexByte(str, '<')])
			inhead = true
			b.WriteString(str)
		} else if strings.TrimSpace(str) == "</head>" {
			b.WriteString(str)
			break
		} else if inhead {
			ctl.AddLabel(strings.TrimSpace(str))
		} else {
			b.WriteString(str)
		}
	}

	_, err = io.Copy(ctl, &b)
	for _, meta := range ctl.Meta {
		fmt.Fprintf(ctl, "%s\t%s\n", space, meta.String())
	}
	fmt.Fprintf(ctl, "%s</head>\n", space)

	// fmt.Printf("struct: %# v\n", pretty.Formatter(ctl.Meta))

	io.Copy(ctl, br)
	return err
}

func (ctl *ControllerStatic) PushAll(path string) error {
	return nil
}

func (ctl *ControllerStatic) NewHTMLHandlerFunc(path string) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		ctl := ctl.Clone()
		ctl.Init(ctx)
		ctl.WriteHTML(path)
	}
}

func (ctl *ControllerStatic) NewStaticHandlerFunc(perfix string) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		path := perfix + ctx.Path()[1:]

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
			eudore.HandlerFile(ctx, path)
		} else {
			ctx.Fatal(err)
		}
	}
}

// func (ctl *ControllerStatic)

func (ctl *ControllerStatic) AddMeta(name, content string) {
	ctl.Meta = append(ctl.Meta, &mapLabel{
		data: fmt.Sprintf("<meta name='%s' content='%s'>", name, content),
	})
}

func (ctl *ControllerStatic) AddLabel(str string) {
	// ctl.Meta = append(ctl.Meta, mapLabel{data: str})
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
	ctl.Hook(ctl, lab)
	ctl.Meta = append(ctl.Meta, lab)

	// fmt.Printf("struct: %# v\n", pretty.Formatter(lab))

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

func (ctl *ControllerStatic) WithPushHook() {
	fn := ctl.Hook
	ctl.Hook = func(ctl *ControllerStatic, lab HeadLabel) {
		fn(ctl, lab)

		var path string
		switch lab.Type() {
		case "script":
			path = lab.(*mapLabel).Get("src")
		case "link":
			path = lab.(*mapLabel).Get("herf")
		default:
			return
		}
		if path != "" {
			ctl.Push(path[1:len(path)-1], nil)
		}
	}
}

func (ctl *ControllerStatic) WithSRIHook() {
	fn := ctl.Hook
	ctl.Hook = func(ctl *ControllerStatic, lab HeadLabel) {
		maplab, _ := lab.(*mapLabel)
		switch lab.Type() {
		case "script":
			sri, err := ctl.GetSha512Value(maplab.keys["src"])
			if err == nil {
				maplab.keys["integrity"] = sri
			} else {
				ctl.Error(err)
			}
		case "link":
			sri, err := ctl.GetSha512Value(maplab.keys["href"])
			if err == nil {
				maplab.keys["integrity"] = sri
			} else {
				ctl.Error(err)
			}
		default:
			fn(ctl, lab)
			return
		}

		fn(ctl, lab)
	}
}

// 计算SRI512
func (ctl *ControllerStatic) GetSha512Value(path string) (string, error) {
	// read
	read, err := ctl.GetPathBody("static" + path[1:len(path)-1])
	if err != nil {
		return "", err
	}

	h := ctl.hashPool.Get().(hash.Hash)
	h.Reset()
	_, err = io.Copy(h, read)

	var val string
	if err == nil {
		val = "sha512-" + base64.StdEncoding.EncodeToString(h.Sum(nil))
	}
	ctl.hashPool.Put(h)
	// read.Close()
	return val, err
}

func (ctl *ControllerStatic) GetPathBody(path string) (io.ReadCloser, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		response, err := http.Get(path)
		return response.Body, err
	}
	// return os.Open(path)
	mf, err := ctl.OpenMergeFile(path)
	return mf, err
}

func (ctl *ControllerStatic) NewMergeFileHandlerFunc(parent string) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		ctl := ctl.Clone()
		ctl.Init(ctx)
		mf, err := ctl.OpenMergeFile(parent + ctl.Path())
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
		h.Set("Last-Modified", mf.modtime.UTC().Format(eudore.TimeFormat))
		h.Set("Content-Type", mf.ext)
		if h.Get("Content-Encoding") == "" {
			h.Set("Content-Length", strconv.FormatInt(mf.length, 10))
		}

		_, err = io.Copy(ctx, mf)
		if err != nil {
			ctx.Fatal((err))
		}
	}
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

func (ctl *ControllerStatic) OpenMergeFile(path string) (*MergeFile, error) {
	fs := &MergeFile{}

	pos := strings.LastIndexByte(path, '/') + 1
	dir, name := path[:pos], path[pos:]

	pos = strings.IndexByte(name, '.')
	ext := ""
	if pos != -1 {
		ext = name[pos:]
		name = name[:pos]
		fs.ext = mime.TypeByExtension(ext)
	}
	paths := strings.Split(name, "-")
	for i := range paths {
		paths[i] = dir + paths[i] + ext
	}

	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			ctl.Error(err)
			fs.Close()
			return nil, err
		}

		desc, err := f.Stat()
		if err != nil {
			ctl.Error(err)
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
	errs := eudore.NewErrors()
	for _, r := range mf.files {
		errs.HandleError(r.Close())
	}
	return errs.GetError()
}
