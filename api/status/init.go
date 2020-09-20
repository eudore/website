package status

import ( 
	"github.com/eudore/website/framework"
)

// Init 函数定义status初始化内容。
func Init(app *framework.App) error {
	api := app.Group("/api/v1/status")
	api.GetFunc("/app", getSystem)
	api.GetFunc("/build", getBuild)
	api.GetFunc("/system", getSystem)
	api.GetFunc("/config", getConfig(app.Config))
	api.GetFunc("/web", getSystem)
	return nil
}
