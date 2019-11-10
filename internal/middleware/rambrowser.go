package middleware

/*
实现ram扩展browser条件检查
*/

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware/ram"

	"github.com/mssola/user_agent"
)

type (
	ConditionBrowser struct {
		Versions map[string]uaVersion
	}
	uaVersion struct {
		min int
		max int
	}
)

func init() {
	ram.RegisterCondition("browser", newConditionBrowser)
}

func newConditionBrowser(i interface{}) ram.Condition {
	cond := &ConditionBrowser{Versions: make(map[string]uaVersion)}
	for _, i := range ram.GetArrayString(i) {
		cond.newBrowser(i)
	}
	return cond
}

func (cond *ConditionBrowser) Name() string {
	return "browser"
}

func (cond *ConditionBrowser) Match(ctx eudore.Context) bool {
	name, version := cond.getBrowser(ctx)
	v := cond.Versions[name]
	if v.min < version && version < v.max {
		return true
	}
	return false
}

func (cond *ConditionBrowser) getBrowser(ctx eudore.Context) (string, int) {
	// 获取browser参数返回版本信息
	browser := ctx.GetParam("browser")
	if browser != "" {
		pos := strings.IndexByte(browser, '/')
		if pos == -1 {
			return browser, 0
		}
		return browser[:pos], eudore.GetStringInt(browser[pos+1:])
	}

	// 分析browser版本并保存
	ua := user_agent.New(ctx.GetHeader("User-Agent"))
	browser, version := ua.Browser()
	if pos := strings.IndexByte(version, '.'); pos != -1 {
		version = version[:pos]
	}
	ctx.SetParam("browser", fmt.Sprintf("%s/%s", browser, version))
	return browser, eudore.GetStringInt(version)
}

func (cond *ConditionBrowser) newBrowser(name string) {
	pos := strings.LastIndexByte(name, '/')
	if pos == -1 {
		cond.Versions[name] = uaVersion{0, 0xff}
		return
	}
	version := name[pos+1:]
	name = name[:pos]
	var min, max = 0, 0xff
	if version[len(version)-1] == '+' {
		min = (eudore.GetStringDefaultInt(version[:len(version)-1], 0))
	} else if pos := strings.Index(version, "-"); pos != -1 {
		min = (eudore.GetStringDefaultInt(version[:pos], 0))
		max = (eudore.GetStringDefaultInt(version[pos+1:], 0xff))
	}
	cond.Versions[name] = uaVersion{min, max}
}
func (cond *ConditionBrowser) MarshalJSON() ([]byte, error) {
	data := make([]string, 0, len(cond.Versions))
	for name, version := range cond.Versions {
		data = append(data, name+version.String())
	}
	return json.Marshal(data)
}

func (ua uaVersion) String() string {
	if ua.min != 0 && ua.max == 0xff {
		return fmt.Sprintf("/%d+", ua.min)
	} else if ua.min != 0 || ua.max != 0xff {
		return fmt.Sprintf("/%d-%d", ua.min, ua.max)
	}
	return ""
}
