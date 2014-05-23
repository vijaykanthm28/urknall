package urknall

type command interface {
	Shell() string   // Used for executing the action in a shell (locally or via ssh).
	Logging() string // Get string used for logging.
}

type renderer interface {
	Render(i interface{})
}