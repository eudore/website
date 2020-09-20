package status

import (
	"github.com/eudore/eudore"
	"runtime"
)

type buildInfo []struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Link    string `json:"link"`
}

func getBuild(ctx eudore.Context) {
	info := buildInfo{
		{Name: "Server Language", Version: runtime.Version(), Link: "https://golang.org"},
		{Name: "Server Web Framework", Version: "eudore", Link: "https://github.com/eudore/eudore"},
		{Name: "Server Database", Version: "PostgreSQL 10.5", Link: "https://www.postgresql.org"},
		{Name: "Server Database Driver", Version: "github.com/lib/pq", Link: "https://github.com/lib/pq"},
		{Name: "Server Captcha", Version: "github.com/dchest/captcha", Link: "https://github.com/dchest/captcha"},
		{Name: "Frotend Web Framework", Version: "Vue.js v2.6.11", Link: "https://github.com/vuejs/vue"},
		{Name: "Frotend Highlighting", Version: "PrismJS 1.15.0", Link: "https://github.com/PrismJS/prism"},
		{Name: "Frotend Markdown", Version: "Marked", Link: "https://github.com/markedjs/marked"},
		{Name: "Frotend Markdown Editor", Version: "SimpleMDE", Link: "https://github.com/sparksuite/simplemde-markdown-editor"},
		{Name: "Frotend Text Editor", Version: "wangEditor", Link: "https://github.com/wangfupeng1988/wangEditor/"},
		// {Name: "", Version: "", Link: ""},
	}
	ctx.Render(info)
}
