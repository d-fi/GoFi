package dfi

func info(message string) string {
	return "ℹ info " + message
}

func warn(message string) string {
	return "⚠ warn " + message
}

func pending(message string) string {
	return "● pending " + message
}

func success(message string) string {
	return "✔ success " + message
}

func failure(message string) string {
	return "✖ error " + message
}

func note(message string) string {
	return "  → " + message
}
