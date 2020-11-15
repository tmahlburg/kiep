package kiep

import "os"

func InstallStatic() {
	destDir := getArchiveDir()
	os.MkdirAll(destDir, os.ModePerm)
}
