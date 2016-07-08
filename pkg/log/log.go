package log

import (
	"fmt"
	"github.com/CodisLabs/codis/pkg/utils/bytesize"
	"github.com/CodisLabs/codis/pkg/utils/log"
	"strings"
)

var (
	maxFileFrag       = 10000000
	maxFragSize int64 = bytesize.GB * 1
)

func SetLogLevel(level string) string {
	level = strings.ToLower(level)
	var l = log.LEVEL_INFO
	switch level {
	case "error":
		l = log.LEVEL_ERROR
	case "warn", "warning":
		l = log.LEVEL_WARN
	case "debug":
		l = log.LEVEL_DEBUG
	case "info":
		fallthrough
	default:
		level = "info"
		l = log.LEVEL_INFO
	}
	log.SetLevel(l)
	log.Infof("set log level to <%s>", level)

	return level
}

func InitLog(file string) {
	// set output log file
	if "" != file {
		f, err := log.NewRollingFile(file, maxFileFrag, maxFragSize)
		if err != nil {
			log.PanicErrorf(err, "open rolling log file failed: %s", file)
		} else {
			log.StdLog = log.New(f, "")
		}
	}

	log.SetLevel(log.LEVEL_INFO)
	log.SetFlags(log.Flags() | log.Lshortfile)
}

func Errorf1(module string, format string, args ...interface{}) {
	log.Errorf(fmt.Sprintf("%s%s", module, format), args...)
}

func Errorln1(module string, args ...interface{}) {
	log.Errorf("%s%s", module, args[0])
}

func Infof1(module string, format string, args ...interface{}) {
	log.Infof(fmt.Sprintf("%s%s", module, format), args...)
}

func Infoln1(module string, args ...interface{}) {
	log.Infof("%s%s", module, args[0])
}

func Warnf1(module string, format string, args ...interface{}) {
	log.Warnf(fmt.Sprintf("%s%s", module, format), args...)
}

func Warnln1(module string, args ...interface{}) {
	log.Warnf("%s%s", module, args[0])
}

func InfoErrorf1(module string, err error, format string, args ...interface{}) {
	log.InfoErrorf(err, fmt.Sprintf("%s%s", module, format), args...)
}

func InfoErrorln1(module string, err error) {
	log.InfoError(err, module)
}

func Debugf1(module string, format string, args ...interface{}) {
	log.Debugf(fmt.Sprintf("%s%s", module, format), args...)
}

func Debugln1(module string, args ...interface{}) {
	log.Debugf("%s%s", module, args[0])
}
