package "kiep"

import "os"

func getArchiveDir() string {
	if dirName := os.Getenv("KIEP_ARCHIVE_DIR"); dirName != "" {
		return dirName
	} else if dirName := path.Join(os.Getenv("XDG_DOCUMENTS_DIR"), "kiep"); dirName != "" {
		return dirName
	} else {
		return path.Join(os.Getenv("HOME"), "Documents/kiep")
	}
}
