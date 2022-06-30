package mqrr

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

var log = logrus.New()

// LogDisableColors disables the color of the log output.
var LogDisableColors = false

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(newTextFormatter())
}

// SetLogWriter sets the default io.Writer used by mqrr for log output.
func SetLogWriter(w io.Writer) {
	log.SetOutput(w)
}

// SetLogFormatter sets the log formatter of the mqrr log output.
func SetLogFormatter(formatter logrus.Formatter) {
	log.SetFormatter(formatter)
}

func debugPrint(format string, values ...interface{}) {
	if IsDebugging() {
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		fmt.Fprintf(log.Out, "[MQRR] "+format, values...)
	}
}

// textFormatter is the default log formatter of mqrr.
type textFormatter logrus.TextFormatter

func newTextFormatter() *textFormatter {
	tf := textFormatter(logrus.TextFormatter{
		FullTimestamp:          true,
		TimestampFormat:        "2006-01-02 15:04:05.000",
		DisableLevelTruncation: true,
	})
	return &tf
}

func (f *textFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = 31 // gray
	case logrus.WarnLevel:
		levelColor = 33 // yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = 31 // red
	default:
		levelColor = 36 // blue
	}
	timestamp := entry.Time.Format(f.TimestampFormat)
	level := strings.ToUpper(entry.Level.String())

	if LogDisableColors {
		return []byte(fmt.Sprintf("%s [%s] [MQRR] %s\n", timestamp, level, entry.Message)), nil
	} else {
		return []byte(fmt.Sprintf("%s \x1b[%dm[%s]\x1b[0m \x1b[36m[MQRR]\x1b[0m %s\n",
			timestamp, levelColor, level, entry.Message)), nil
	}
}
