package main

import (
	"github.com/eudore/website/api/auth"
	"github.com/eudore/website/api/note"
	"github.com/eudore/website/api/status"
	"github.com/eudore/website/api/term"
	"github.com/eudore/website/framework"
)

func main() {
	framework.New().Run(
		auth.Init,
		note.Init,
		term.Init,
		status.Init,
	)
}
