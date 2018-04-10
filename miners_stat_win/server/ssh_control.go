package main

// link: http://blog.ralch.com/tutorial/golang-ssh-connection/
import (
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	CERT_PASSWORD        = 1
	CERT_PUBLIC_KEY_FILE = 2
	DEFAULT_TIMEOUT      = 3 // second
)

type SSH struct {
	Ip      string
	User    string
	Cert    string //password or key file path
	Port    int
	session *ssh.Session
	client  *ssh.Client
}

func (ssh_client *SSH) readPublicKeyFile(file string) ssh.AuthMethod {
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

func (ssh_client *SSH) Connect(mode int) error {

	var ssh_config *ssh.ClientConfig
	var auth []ssh.AuthMethod
	if mode == CERT_PASSWORD {
		auth = []ssh.AuthMethod{ssh.Password(ssh_client.Cert)}
	} else if mode == CERT_PUBLIC_KEY_FILE {
		auth = []ssh.AuthMethod{ssh_client.readPublicKeyFile(ssh_client.Cert)}
	} else {
		return nil
	}

	ssh_config = &ssh.ClientConfig{
		User: ssh_client.User,
		Auth: auth,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: time.Second * DEFAULT_TIMEOUT,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ssh_client.Ip, ssh_client.Port), ssh_config)
	if err != nil {
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return err
	}

	ssh_client.session = session
	ssh_client.client = client
	return nil
}

func (ssh_client *SSH) RunCmd(cmd string) error {
	_, err := ssh_client.session.CombinedOutput(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (ssh_client *SSH) Close() error {
	err := ssh_client.session.Close()
	if err != nil {
		return err
	}
	err = ssh_client.client.Close()
	if err != nil {
		return err
	}
	return nil
}

//demo
func Reboot(ip string) error {
	client := &SSH{
		Ip:   ip,
		User: "ethos",
		Port: 22,
		Cert: "Your ethos passwd",
	}

	err := client.Connect(CERT_PASSWORD)
	if err != nil {
		return err
	}
	err = client.RunCmd("r")
	if err != nil {
		return err
	}
	err = client.Close()
	if err != nil {
		return err
	}
	return nil
}
