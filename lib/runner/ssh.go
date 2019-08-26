package runner

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/axetroy/go-fs"
	"github.com/axetroy/sshunter/lib/parser"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"os"
	"path"
	"strings"
)

type Client struct {
	config     parser.Config
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

func NewSSH(c parser.Config) *Client {
	return &Client{
		config:     c,
		sshClient:  nil,
		sftpClient: nil,
	}
}

func (c *Client) Connect() error {
	// connect to
	sshConfig := &ssh.ClientConfig{
		User: c.config.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.config.Password),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	addr := fmt.Sprintf("%s:%v", c.config.Host, c.config.Port)

	sshClient, err := ssh.Dial("tcp", addr, sshConfig)

	if err != nil {
		return err
	}

	c.sshClient = sshClient

	// create sftp client
	sftpClient, err := sftp.NewClient(sshClient)

	if err != nil {
		return err
	}

	c.sftpClient = sftpClient

	return nil
}

func (c *Client) Disconnect() error {
	sshErr := c.sshClient.Close()
	sftpErr := c.sftpClient.Close()

	if sshErr != nil {
		return sshErr
	}

	if sftpErr != nil {
		return sftpErr
	}

	return nil
}

func (c *Client) Pwd() (string, error) {
	// Create a session. It is one session per command.
	session, err := c.sshClient.NewSession()

	if err != nil {
		return "", err
	}

	defer func() {
		_ = session.Close()
	}()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err = session.Run("pwd"); err != nil {
		msg := stderr.String()

		if len(msg) > 0 {
			return "", errors.New(msg)
		}

		return "", err
	}

	return strings.Trim(strings.Trim(stdout.String(), " "), "\n"), nil
}

func (c *Client) Run(command string) error {
	// Create a session. It is one session per command.
	session, err := c.sshClient.NewSession()

	if err != nil {
		return err
	}

	defer func() {
		_ = session.Close()
	}()

	sessionStdOut, err := session.StdoutPipe()

	if err != nil {
		return err
	}

	go io.Copy(os.Stdout, sessionStdOut)

	sessionStdErr, err := session.StderrPipe()

	if err != nil {
		return err
	}

	go io.Copy(os.Stdout, sessionStdErr)

	if c.config.CWD != "" {
		command = "cd " + c.config.CWD + " && " + command
	}

	var setEnvCommand []string

	// set environmental variable before run
	for key, value := range c.config.Env {
		// export KEY=VALUE
		setEnvCommand = append(setEnvCommand, fmt.Sprintf("export %s=%s", key, value))
	}

	if len(setEnvCommand) != 0 {
		command = strings.Join(setEnvCommand, " && ") + " && " + command
	}

	if err = session.Run(command); err != nil {
		return err
	}

	return nil
}

func (c *Client) downloadFile(remoteFilePath string, localDir string) error {
	remoteFile, err := c.sftpClient.Open(remoteFilePath)

	if err != nil {
		return err
	}

	defer remoteFile.Close()

	var localFileName = path.Base(remoteFilePath)

	// ensure local dir exist
	if err := fs.EnsureDir(localDir); err != nil {
		return err
	}

	localFile, err := os.Create(path.Join(localDir, localFileName))

	if err != nil {
		return err
	}

	defer localFile.Close()

	if _, err = remoteFile.WriteTo(localFile); err != nil {
		return err
	}

	return nil
}

func (c *Client) downloadDir(remoteFilePath string, localDir string) error {
	files, err := c.sftpClient.ReadDir(remoteFilePath)
	if err != nil {
		return err
	}

	localDir = path.Join(localDir, path.Base(remoteFilePath))

	for _, file := range files {
		fileName := file.Name()
		absFilePath := path.Join(remoteFilePath, fileName)

		if file.IsDir() {
			if err := c.downloadDir(absFilePath, path.Join(localDir, fileName)); err != nil {
				return nil
			}
		} else {
			if err := c.downloadFile(absFilePath, localDir); err != nil {
				return nil
			}
		}
	}

	return nil
}

func (c *Client) Download(remoteFilePath string, localDir string) error {
	remoteFileStat, err := c.sftpClient.Stat(remoteFilePath)

	if err != nil {
		return err
	}

	// if it is a directory
	if remoteFileStat.IsDir() {
		return c.downloadDir(remoteFilePath, localDir)
	} else {
		return c.downloadFile(remoteFilePath, localDir)
	}
}

func (c *Client) Copy(localFilePath string, remoteDir string) error {
	if err := c.Run(fmt.Sprintf("mkdir -p %s", remoteDir)); err != nil {
		return err
	}

	localFile, err := os.Open(localFilePath)

	if err != nil {
		return err
	}

	defer localFile.Close()

	remoteFileName := path.Base(localFilePath)

	dstFile, err := c.sftpClient.Create(path.Join(remoteDir, remoteFileName))

	if err != nil {
		return err
	}

	defer dstFile.Close()

	buf := make([]byte, 1024)
	for {
		n, err := localFile.Read(buf)

		if err != nil {
			return err
		}

		if n == 0 {
			break
		}
		if _, err := dstFile.Write(buf[0:n]); err != nil {
			return err
		}
	}

	return nil
}
