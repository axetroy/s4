package ssh

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

type Options struct {
	CWD string            `json:"cwd"`
	Env map[string]string `json:"env"`
}

var (
	// Linux 的内置目录路径，删除这些路径可能会导致系统崩溃
	// 可以删除他下面的路径，但是不能直接删除目录
	linuxBuildInPathMap = map[string]int{
		"":            1,
		".":           1,
		"/":           1,
		"/bin":        1,
		"/boot":       1,
		"/dev":        1,
		"/etc":        1,
		"/home":       1,
		"/lib":        1,
		"/lost+found": 1,
		"/media":      1,
		"/mnt":        1,
		"/opt":        1,
		"/proc":       1,
		"/root":       1,
		"/sbin":       1,
		"/selinux":    1,
		"/srv":        1,
		"/sys":        1,
		"/tmp":        1,
		"/usr":        1,
		"/usr/bin":    1,
		"/usr/sbin":   1,
		"/usr/src":    1,
		"/var":        1,
	}
)

// 判断是否是 linux 的危险路径，通常这个路径是不能删除的
func IsLinuxBuildInPath(filepath string) bool {
	if _, ok := linuxBuildInPathMap[filepath]; ok {
		return true
	} else {
		return false
	}
}

func setEnvForCommand(command string, env map[string]string) (newCommand string) {
	var setEnvCommand []string

	for key, value := range env {
		// export KEY=VALUE
		setEnvCommand = append(setEnvCommand, fmt.Sprintf("export %s=%s;", key, value))
	}

	if len(setEnvCommand) != 0 {
		newCommand = strings.Join(setEnvCommand, " ") + " " + command
	} else {
		newCommand = command
	}

	return
}

func NewSSH() *Client {
	return &Client{
		sshClient:  nil,
		sftpClient: nil,
	}
}

func (c *Client) Connect(host, port, username, password string) error {
	// connect to
	sshConfig := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	addr := fmt.Sprintf("%s:%v", host, port)

	if sshClient, err := ssh.Dial("tcp", addr, sshConfig); err != nil {
		return err
	} else {
		c.sshClient = sshClient

		// create sftp client
		if sftpClient, err := sftp.NewClient(sshClient); err != nil {
			return err
		} else {
			c.sftpClient = sftpClient
		}
	}

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
	return c.sftpClient.Getwd()
}

func (c *Client) Env(key string, options Options) (string, error) {
	command := fmt.Sprintf("echo $%s", key)

	// Create a session. It is one session per command.
	session, err := c.sshClient.NewSession()

	if err != nil {
		return "", err
	}

	defer session.Close()

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	command = setEnvForCommand(command, options.Env)

	if err = session.Run(command); err != nil {
		return "", err
	}

	return strings.TrimSpace(stdoutBuf.String()), nil
}

func (c *Client) RunWithCustomIO(command string, options Options, stdout *bytes.Buffer, stderr *bytes.Buffer) error {
	// Create a session. It is one session per command.
	session, err := c.sshClient.NewSession()

	if err != nil {
		return err
	}

	defer session.Close()

	session.Stdout = stdout
	session.Stderr = stderr

	if options.CWD != "" {
		command = "cd " + options.CWD + " && " + command
	}

	command = setEnvForCommand(command, options.Env)

	if err = session.Run(command); err != nil {
		return err
	}

	return nil
}

// TODO: not use yet
func (c *Client) RunWithCommandPipe(command string, commands []string, options Options) (err error) {
	// Create a session. It is one session per command.
	session, err := c.sshClient.NewSession()

	if err != nil {
		return err
	}

	defer session.Close()

	cmd := exec.Command(commands[0], commands[1:]...)

	if sessionStdOut, err := session.StdoutPipe(); err != nil {
		return err
	} else {
		if cmdStdin, err := cmd.StdinPipe(); err != nil {
			return err
		} else {
			go io.Copy(cmdStdin, sessionStdOut)
		}
	}

	if sessionStdErr, err := session.StderrPipe(); err != nil {
		return err
	} else {
		if cmdStdin, err := cmd.StdinPipe(); err != nil {
			return err
		} else {
			go io.Copy(cmdStdin, sessionStdErr)
		}
	}

	if options.CWD != "" {
		command = "cd " + options.CWD + " && " + command
	}

	command = setEnvForCommand(command, options.Env)

	go func() {
		err = session.Run(command)
	}()

	if err := cmd.Start(); err != nil {
		return err
	}

	return nil
}

func (c *Client) Run(command string, options Options) error {
	// Create a session. It is one session per command.
	session, err := c.sshClient.NewSession()

	if err != nil {
		return err
	}

	defer session.Close()

	if sessionStdOut, err := session.StdoutPipe(); err != nil {
		return err
	} else {
		go io.Copy(os.Stdout, sessionStdOut)
	}

	if sessionStdErr, err := session.StderrPipe(); err != nil {
		return err
	} else {
		go io.Copy(os.Stdout, sessionStdErr)
	}

	if options.CWD != "" {
		command = "cd " + options.CWD + " && " + command
	}

	command = setEnvForCommand(command, options.Env)

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

	remoteFileStat, err := remoteFile.Stat()

	if err != nil {
		return err
	}

	localFileName := path.Base(remoteFilePath)
	localFilePath := path.Join(localDir, localFileName)

	// ensure local dir exist
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return err
	}

	localFile, err := os.Create(localFilePath)

	if err != nil {
		return err
	}

	defer localFile.Close()

	remoteFileSize := remoteFileStat.Size()

	tmpl := fmt.Sprintf(`{{string . "prefix"}}{{ green "%s" }} {{counters . }} {{ bar . "[" "=" ">" "-" "]"}} {{percent . }} {{speed . }}{{string . "suffix"}}`, localFilePath)

	// start bar based on our template
	bar := pb.ProgressBarTemplate(tmpl).Start64(remoteFileSize)

	bar.Set(pb.Bytes, true)
	bar.SetWriter(os.Stdout)

	barReader := bar.NewProxyReader(remoteFile)

	// update mode
	if err := os.Chmod(localFilePath, remoteFileStat.Mode()); err != nil {
		return err
	}

	if _, err := io.Copy(localFile, barReader); err != nil {
		return err
	}

	bar.Finish()

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

func (c *Client) uploadFile(localFilePath string, remoteDir string) error {
	localFile, err := os.Open(localFilePath)

	if err != nil {
		return err
	}

	defer localFile.Close()

	localFileStat, err := localFile.Stat()

	if err != nil {
		return err
	}

	remoteFileName := path.Base(localFilePath)
	remoteFilePath := path.Join(remoteDir, remoteFileName)

	if err := c.sftpClient.MkdirAll(remoteDir); err != nil {
		return err
	}

	remoteFile, err := c.sftpClient.Create(remoteFilePath)

	if err != nil {
		return err
	}

	defer remoteFile.Close()

	// update file mode
	if err := c.sftpClient.Chmod(remoteFilePath, localFileStat.Mode()); err != nil {
		return err
	}

	localFileSize := localFileStat.Size()

	tmpl := fmt.Sprintf(`{{string . "prefix"}}{{ green "%s" }} {{counters . }} {{ bar . "[" "=" ">" "-" "]"}} {{percent . }} {{speed . }}{{string . "suffix"}}`, remoteFilePath)

	// start bar based on our template
	bar := pb.ProgressBarTemplate(tmpl).Start64(localFileSize)

	bar.Set(pb.Bytes, true)
	bar.SetWriter(os.Stdout)

	localFileReader := bufio.NewReader(localFile)

	barReader := bar.NewProxyReader(localFileReader)

	if _, err := remoteFile.ReadFrom(barReader); err != nil {
		return err
	}

	bar.Finish()

	return nil
}

func (c *Client) uploadDir(localFilePath string, remoteDir string) error {
	files, err := ioutil.ReadDir(localFilePath)

	if err != nil {
		return err
	}

	remoteDir = path.Join(remoteDir, path.Base(localFilePath))

	for _, file := range files {
		fileName := file.Name()
		absFilePath := path.Join(localFilePath, fileName)
		if file.IsDir() {
			if err = c.uploadDir(absFilePath, remoteDir); err != nil {
				return err
			}
		} else {
			if err := c.uploadFile(absFilePath, remoteDir); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Client) Upload(localFilePath string, remoteDir string) error {
	localStat, err := os.Stat(localFilePath)

	if err != nil {
		return err
	}

	if localStat.IsDir() {
		return c.uploadDir(localFilePath, remoteDir)
	} else {
		return c.uploadFile(localFilePath, remoteDir)
	}
}

func (c *Client) Copy(sourceFilepath string, destinationFilepath string) error {
	sourceFile, err := c.sftpClient.Open(sourceFilepath)

	if err != nil {
		return err
	}

	defer sourceFile.Close()

	destinationFile, err := c.sftpClient.Create(destinationFilepath)

	if err != nil {
		return err
	}

	defer destinationFile.Close()

	if _, err = sourceFile.WriteTo(destinationFile); err != nil {
		return err
	}

	// update new file mode and time
	if sourceFileStat, err := sourceFile.Stat(); err != nil {
		return err
	} else {
		err := c.sftpClient.Chmod(destinationFilepath, sourceFileStat.Mode())

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) Move(oldFilepath string, newFilepath string) error {
	return c.sftpClient.Rename(oldFilepath, newFilepath)
}

func (c *Client) Delete(files ...string) error {
	for _, file := range files {
		// Prevent the removal of dangerous system files
		if IsLinuxBuildInPath(file) {
			fmt.Printf("Prevent the removal of dangerous system file '%s'\n", file)
			continue
		}

		// if path not exist. then ignore error
		if _, err := c.sftpClient.Stat(file); err != nil {
			continue
		}

		if err := c.sftpClient.Remove(file); err != nil {
			return err
		}
	}

	return nil
}
