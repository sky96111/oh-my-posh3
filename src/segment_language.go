package main

type loadContext func()

type inContext func() bool

type language struct {
	props        *properties
	env          environmentInfo
	extensions   []string
	commands     []string
	executable   string
	versionParam string
	versionRegex string
	version      string
	exitCode     int
	loadContext  loadContext
	inContext    inContext
}

const (
	// DisplayMode sets the display mode (always, when_in_context, never)
	DisplayMode Property = "display_mode"
	// DisplayModeAlways displays the segment always
	DisplayModeAlways string = "always"
	// DisplayModeFiles displays the segment when the current folder contains certain extensions
	DisplayModeFiles string = "files"
	// DisplayModeEnvironment displays the segment when the environment has a language's context
	DisplayModeEnvironment string = "environment"
	// DisplayModeContext displays the segment when the environment or files is active
	DisplayModeContext string = "context"
	// MissingCommandTextProperty sets the text to display when the command is not present in the system
	MissingCommandTextProperty Property = "missing_command_text"
	// MissingCommandText displays empty string by default
	MissingCommandText string = ""
)

func (l *language) string() string {
	if !l.props.getBool(DisplayVersion, true) {
		return ""
	}
	if !l.hasCommand() {
		return l.props.getString(MissingCommandTextProperty, MissingCommandText)
	}
	l.setVersion()
	return l.version
}

func (l *language) enabled() bool {
	l.loadLanguageContext()
	displayMode := l.props.getString(DisplayMode, DisplayModeFiles)
	switch displayMode {
	case DisplayModeAlways:
		return true
	case DisplayModeEnvironment:
		return l.inLanguageContext()
	case DisplayModeFiles:
		return l.hasLanguageFiles()
	case DisplayModeContext:
		fallthrough
	default:
		return l.hasLanguageFiles() || l.inLanguageContext()
	}
}

// hasLanguageFiles will return true at least one file matching the extensions is found
func (l *language) hasLanguageFiles() bool {
	for i, extension := range l.extensions {
		if l.env.hasFiles(extension) {
			break
		}
		if i == len(l.extensions)-1 {
			return false
		}
	}

	return true
}

// getVersion returns the version and exit code returned by the executable
func (l *language) setVersion() {
	versionInfo, err := l.env.runCommand(l.executable, l.versionParam)
	if exitErr, ok := err.(*commandError); ok {
		l.exitCode = exitErr.exitCode
		return
	}
	values := findNamedRegexMatch(l.versionRegex, versionInfo)
	l.exitCode = 0
	l.version = values["version"]
}

// hasCommand checks if one of the commands exists and sets it as executable
func (l *language) hasCommand() bool {
	for i, command := range l.commands {
		if l.env.hasCommand(command) {
			l.executable = command
			break
		}
		if i == len(l.commands)-1 {
			return false
		}
	}
	return true
}

func (l *language) loadLanguageContext() {
	if l.loadContext == nil {
		return
	}
	l.loadContext()
}

func (l *language) inLanguageContext() bool {
	if l.inContext == nil {
		return false
	}
	return l.inContext()
}
