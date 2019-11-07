package util

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
