package task

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type (
	AgentConfig struct {
		Name  string `json:"name,omitempty"`
		Addr  string `json:"addr"`
		Uri   string `json:"uri"`
		Toekn string `json:"token"`
	}
	AgentExecutor struct {
		client *http.Client
		Addr   string `json:"addr"`
		Uri    string `json:"uri"`
		Toekn  string `json:"token"`
	}
)

func (ac *AgentConfig) EncodeConfig() string {
	name := ac.Name
	ac.Name = ""
	data, err := json.Marshal(ac)
	ac.Name = name
	if err != nil {
		return "{}"
	}
	return string(data)
}

func NewAgentExecutor(config string) *AgentExecutor {
	agent := &AgentExecutor{client: http.DefaultClient}
	json.Unmarshal([]byte(config), agent)
	return agent
}

func (exec *AgentExecutor) Run(task *Task) error {
	req, err := http.NewRequest("PUT", fmt.Sprintf("http://127.0.0.1:9090%s?token=%s", exec.Uri, exec.Toekn), strings.NewReader(task.Params["command"].(string)))
	if err != nil {
		fmt.Println(err)
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
