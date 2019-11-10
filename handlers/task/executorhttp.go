package task

import (
	"io/ioutil"
	"net/http"
)

type HttpExecutor struct {
	client *http.Client
}

func NewHttpExecutor() Executor {
	return &HttpExecutor{client: http.DefaultClient}
}

func (exec *HttpExecutor) Run(task *Task) error {
	data := task.Params
	req, err := http.NewRequest(data["method"].(string), data["url"].(string), nil)
	if err != nil {
		return err
	}

	req = req.WithContext(task.Context)
	resp, err := exec.client.Do(req)
	if err != nil {
		return err
	}

	str, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	task.Message = str
	return err
}
