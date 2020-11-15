package main

import (
	"os"

	"github.com/tmahlburg/kiep/internal/app/kiep"
)

func main() {
	// archival type
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "article":
			if len(os.Args) > 2 {
				kiep.ArchiveArticle(os.Args[2], os.Args[3:])
			} else {
				kiep.PrintHelp()
				os.Exit(1)
			}
		case "init":
			kiep.InstallStatic()
		case "help":
			kiep.PrintHelp()
		default:
			kiep.PrintHelp()
			os.Exit(1)
		}
	} else {
		kiep.PrintHelp()
		os.Exit(1)
	}
}
