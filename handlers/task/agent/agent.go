package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type Midd struct {
	Token string
}

func (m *Midd) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	token := r.Form.Get("token")
	if token != m.Token {
		http.Error(w, http.StatusText(403), 403)
	} else {
		http.DefaultServeMux.ServeHTTP(w, r)
	}

}

func main() {
	http.HandleFunc("/getfile/", handlegetfile)
	http.HandleFunc("/upfile/", handleupfile)
	http.HandleFunc("/bash", handlebash)
	http.HandleFunc("/check", handlecheck)
	token := getRandomString()
	go func() {
		addr := os.Getenv("ENV_AGENT_SERVER")
		if len(addr) != 0 {
			// addr = fmt.Sprintf("%s?server=http://%s:9090/bash/%%3Ftoken=%s", addr, getaddr(addr), token)
			log.Printf("agent server register addr: %s\n", addr)
			sendtoken(addr, token)
			for range time.NewTicker(10 * time.Second).C {
				sendtoken(addr, token)
			}
		}
	}()
	log.Println("start server access token is", token)
	http.ListenAndServe(":9090", &Midd{token})
}

func sendtoken(addr, token string) error {
	data := map[string]interface{}{
		"name":  "sh-agent",
		"uri":   "/bash",
		"token": token,
	}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", addr, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("agent sendtoken response status is %d", resp.StatusCode)
	}
	return nil
}

func getRandomString() string {
	const letters = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
	result := make([]byte, 16)
	for i := range result {
		result[i] = letters[rand.Intn(61)]
	}
	return string(result)
}

func handlegetfile(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[8:])
}

func handlebash(w http.ResponseWriter, r *http.Request) {
	err := execBash(r.Context(), w, w, r.Body)
	if err != nil {
		log.Println(err)
	}
}

func execBash(ctx context.Context, stdout, stderr io.Writer, body io.Reader) error {
	filename := "/tmp/0000.sh"
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer os.Remove(filename)
	file.WriteString("#!/bin/bash\n")
	// file.WriteString("set -v\n")
	io.Copy(file, body)
	file.Close()

	return execCommand(ctx, stdout, stderr, "/bin/bash", filename)
}

func execCommand(ctx context.Context, stdout, stderr io.Writer, name string, params ...string) error {
	cmd := exec.CommandContext(ctx, name, params...)
	cmd.Dir = "/tmp"
	if stdout != nil {
		cmd.Stdout = stdout
	}
	if stderr != nil {
		cmd.Stderr = stderr
	}
	return cmd.Run()
}

func handleupfile(w http.ResponseWriter, r *http.Request) {
	file, err := os.OpenFile(r.URL.Path[7:], os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	io.Copy(file, r.Body)
	file.Close()
}
func handlecheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}
