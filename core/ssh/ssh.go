	"github.com/axetroy/s4/core/util"
		// Prevent the removal of dangerous system files
		if util.IsLinuxBuildInPath(file) {
			fmt.Printf("Prevent the removal of dangerous system file '%s'\n", file)
		// if path not exist. then ignore error
		if _, err := c.sftpClient.Stat(file); err != nil {
		if err := c.sftpClient.Remove(file); err != nil {