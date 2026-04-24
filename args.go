package main

// argsContainHelpFlag reports whether argv includes a help flag (so we can skip .env).
func argsContainHelpFlag(argv []string) bool {
	if len(argv) == 0 {
		return false
	}
	for _, a := range argv {
		switch a {
		case "-h", "-help", "--help":
			return true
		}
	}
	return false
}
