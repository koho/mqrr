package mqrr

const (
	DebugMode   = "debug"
	ReleaseMode = "release"
)

const (
	debugCode = iota
	releaseCode
)

var mode = debugCode

// SetMode sets running mode according to input string.
func SetMode(value string) {
	if value == "" {
		value = DebugMode
	}

	switch value {
	case DebugMode:
		mode = debugCode
	case ReleaseMode:
		mode = releaseCode
	default:
		panic("mode unknown: " + value)
	}
}

// IsDebugging returns true if the framework is running in debug mode.
// Use SetMode(mqrr.ReleaseMode) to disable debug mode.
func IsDebugging() bool {
	return mode == debugCode
}
