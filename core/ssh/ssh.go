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
func (c *Client) Connect(host, port, username, password string) error {
		User: username,
			ssh.Password(password),
	addr := fmt.Sprintf("%s:%v", host, port)
	if sshClient, err := ssh.Dial("tcp", addr, sshConfig); err != nil {
	} else {
		c.sshClient = sshClient
		// create sftp client
		if sftpClient, err := sftp.NewClient(sshClient); err != nil {
			return err
		} else {
			c.sftpClient = sftpClient
		}
func (c *Client) Env(key string, options Options) (string, error) {
	command = setEnvForCommand(command, options.Env)
	if err = session.Run(command); err != nil {
		return "", err
	return strings.TrimSpace(stdoutBuf.String()), nil
}

func (c *Client) RunAndCombineOutput(command string, options Options) ([]byte, error) {
	// Create a session. It is one session per command.
	session, err := c.sshClient.NewSession()

	if err != nil {
		return nil, err
	defer session.Close()

	if options.CWD != "" {
		command = "cd " + options.CWD + " && " + command
	command = setEnvForCommand(command, options.Env)

	return session.CombinedOutput(command)
func (c *Client) RunWithCustomIO(command string, options Options, stdout *bytes.Buffer, stderr *bytes.Buffer) error {
	if options.CWD != "" {
		command = "cd " + options.CWD + " && " + command
	command = setEnvForCommand(command, options.Env)
func (c *Client) Run(command string, options Options) error {
	defer session.Close()

	if options.CWD != "" {
		command = "cd " + options.CWD + " && " + command
	}

	command = setEnvForCommand(command, options.Env)
	if err = session.Run(command); err != nil {
		return err
	if _, err := io.Copy(os.Stdout, sessionStdOut); err != nil {
		return err
	if _, err := io.Copy(os.Stderr, sessionStdErr); err != nil {
		if IsLinuxBuildInPath(file) {