// 基于ssh方式远程执行命令
package task

import (
	// "io"
	// "fmt"
	// "net"
	"context"
	// "time"
	// "io/ioutil"
	"golang.org/x/crypto/ssh"
)

type SSHExecutor struct{}

func (exec *SSHExecutor) Run(ctx context.Context) error {
	return nil
}

type (
	SshExecutor struct {
		Ip      string
		User    string
		Cert    string //password or key file path
		Port    int
		session *ssh.Session
		client  *ssh.Client
	}
)

/*

func NewSshExecutor(ip , name, pass string) (Execer, error) {
	if len(pass) ==0 && name == "root" {
		pass = hosts[ip]
	}
	return &SshExecutor{
		Ip: ip,
		User : name,
		Port: 22,
		Cert: pass,
	}, nil
}



func (ssh_client *SshExecutor) readPublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func (ssh_client *SshExecutor) Connect(mode int) {

	var ssh_config *ssh.ClientConfig
	var auth  []ssh.AuthMethod
	if mode == CERT_PASSWORD {
		auth = []ssh.AuthMethod{ssh.Password(ssh_client.Cert)}
	} else if mode == CERT_PUBLIC_KEY_FILE {
		auth = []ssh.AuthMethod{ssh_client.readPublicKeyFile(ssh_client.Cert)}
	} else {
		fmt.Println("does not support mode: ", mode)
		return
	}

	ssh_config = &ssh.ClientConfig{
		User: ssh_client.User,
		Auth: auth,
		//需要验证服务端，不做验证返回nil就可以，点击HostKeyCallback看源码就知道了
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout:time.Second * 30,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ssh_client.Ip, ssh_client.Port), ssh_config)
	if err != nil {
		fmt.Println(err)
		return
	}

	session, err := client.NewSession()
	if err != nil {
		fmt.Println(err)
		client.Close()
		return
	}

	ssh_client.session = session
	ssh_client.client  = client
}


func (ssh_client *SshExecutor) Close() {
	ssh_client.session.Close()
	ssh_client.client.Close()
}


func (ssh_client *SshExecutor) Exec(cmd string, stdout io.Writer) error {
	ssh_client.Connect(CERT_PASSWORD)
	if ssh_client.session == nil {
		return fmt.Errorf("nil ssh session")
	}

	defer ssh_client.Close()
	ssh_client.session.Stdout = stdout
	ssh_client.session.Stderr = stdout
	return ssh_client.session.Run(cmd)
}




var hosts map[string]string = map[string]string{
	"18.16.200.10": "1qaz2wsxzsxc",
}
*/
