package auth

import (
	"crypto/md5"
	"fmt"
	"github.com/eudore/website/framework"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type IconController struct {
	framework.ControllerWebsite
	IconTemp string
}

type iconInfo struct {
	Icon []byte `alias:"icon"`
	Mail string `alias:"mail"`
}

func NewIconController(app *framework.App) *IconController {
	iconTemp := app.Config.Auth.IconTemp
	if iconTemp != "" {
		os.MkdirAll(filepath.Join(iconTemp, "id"), 0755)
		os.MkdirAll(filepath.Join(iconTemp, "name"), 0755)
	}
	return &IconController{
		ControllerWebsite: framework.ControllerWebsite{
			Context: framework.Context{
				DB: app.DB,
			},
		},
		IconTemp: iconTemp,
	}
}
func (ctl *IconController) GetRouteParam(pkg, name, method string) string {
	if strings.HasPrefix(method, "Get") {
		return ""
	}
	return ctl.ControllerWebsite.GetRouteParam(pkg, name, method)
}

func (ctl *IconController) GetById() error {
	return ctl.loadIconFile("id")
}

func (ctl *IconController) GetNameByName() error {
	return ctl.loadIconFile("name")
}

func (ctl *IconController) loadIconFile(key string) error {
	iconpath := filepath.Join(ctl.IconTemp, key, ctl.GetParam(key)+".png")
	file, err := os.Open(iconpath)
	if err != nil {
		err = ctl.loadIconData(key, iconpath)
		if err != nil {
			return err
		}
		file, err = os.Open(iconpath)
		if err != nil {
			return err
		}
	}
	ctl.SetHeader("Content-Type", "image/png")
	io.Copy(ctl, file)
	return file.Close()
}

func (ctl *IconController) loadIconData(key, iconpath string) error {
	ctl.Debug("loadIconData", key, iconpath)
	var icon []byte
	var mail string
	err := ctl.QueryRow("SELECT mail,icon FROM tb_auth_user_info WHERE "+key+"=$1", ctl.GetParam(key)).Scan(&mail, &icon)
	if err != nil {
		return err
	}
	if len(icon) == 0 {
		return ctl.loadGravatar(iconpath, mail)
	}
	if len(mail) == 0 {
		// copy favicon
	}

	file, err := os.Create(iconpath)
	if err != nil {
		return err
	}
	file.Write(icon)
	return file.Close()
}

func (ctl *IconController) loadGravatar(path, mail string) error {
	hash := md5.Sum([]byte(mail))
	ctl.Debug("loadGravatar", path, mail, fmt.Sprintf("https://www.gravatar.com/avatar/%x?s=%d&d=identicon", hash, 64))
	resp, err := http.Get(fmt.Sprintf("https://www.gravatar.com/avatar/%x?s=%d&d=identicon", hash, 64))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	link := resp.Header.Get("Link")
	if link != "" {
		ctl.SetHeader("Link", link)
	}

	newFile, err := os.Create(path)
	if err != nil {
		return err
	}

	defer newFile.Close()
	_, err = io.Copy(newFile, resp.Body)
	return err
}

func (ctl *IconController) PostByUserid() {
	fileheader := ctl.FormFile("file")
	ctl.Debugf("%s %s %d", fileheader.Filename, fileheader.Header, fileheader.Size)
	ctl.Exec("UPDATE tb_auth_user_info SET icon=$1 WHERE id=$2", "", ctl.GetParam("userid"))
}

func (ctl *IconController) PutByUserid() {
	ctl.Exec("UPDATE tb_auth_user_info SET icon='' WHERE id=$1", ctl.GetParam("userid"))
}
