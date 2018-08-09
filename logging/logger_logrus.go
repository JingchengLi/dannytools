package logging

import (
	"dannytools/constvar"
	"dannytools/ehand"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"sync"

	myLogger "github.com/sirupsen/logrus"
	kitsFile "github.com/toolkits/file"
	kitsSlice "github.com/toolkits/slice"
	//"github.com/davecgh/go-spew/spew"
)

const (
	DEBUG   = "debug"
	INFO    = "info"
	WARNING = "warning"
	ERROR   = "error"
	//CRITICAL = "critical"
)

var (
	LogLevelList []string = []string{
		DEBUG,
		INFO,
		WARNING,
		ERROR,
		//CRITICAL,
	}
	LogLevelToLogrusLevel map[string]myLogger.Level = map[string]myLogger.Level{
		DEBUG:   myLogger.DebugLevel,
		INFO:    myLogger.InfoLevel,
		WARNING: myLogger.WarnLevel,
		ERROR:   myLogger.ErrorLevel,
	}
)

type LogConf struct {
	LogFile   string `mapstructure:"logfile"`
	LogLevel  string `mapstructure:"loglevel"`
	LogFormat string `mapstructure:"logformat"`
}

func CheckLogLevel(lv string) bool {
	return kitsSlice.ContainsString(LogLevelList, lv)
}

func GetLogrusLogLevel(lv string) myLogger.Level {
	if !CheckLogLevel(lv) {
		lv = INFO
	}
	return LogLevelToLogrusLevel[lv]
}

func GetAllLogLevelsString(sep string) string {
	return strings.Join(LogLevelList, sep)
}

func SetLogLevel(lg *myLogger.Logger, lv string) {

	lg.Level = GetLogrusLogLevel(lv)
}

/*
use before we get logging config
*/

func NewRawLogger(lv string) *myLogger.Logger {
	lg := myLogger.New()
	lg.Level = GetLogrusLogLevel(lv)
	lg.Formatter = GetTextFormat()
	//lg.Formatter = myLogger.TextFormatter{ForceColors: false, DisableColors: true, DisableTimestamp: false, TimestampFormat: constvar.DATETIME_FORMAT_NOSPACE}
	return lg
}

func GetTextFormat() *myLogger.TextFormatter {
	return &myLogger.TextFormatter{ForceColors: false, DisableColors: true, DisableTimestamp: false, TimestampFormat: constvar.DATETIME_FORMAT_NOSPACE}
}

func GetJsonFormat() *myLogger.JSONFormatter {
	return &myLogger.JSONFormatter{TimestampFormat: constvar.DATETIME_FORMAT_NOSPACE, DisableTimestamp: false}
}

/*
logFile: full path of log file. default, os.Stdout is used
logLevel: debug, info, warning, error. default warning is used
format: json, text
*/
func (logCf *LogConf) CreateNewLogger() *myLogger.Logger {

	oneLogger := &myLogger.Logger{}

	if logCf.LogFile == "" {
		oneLogger.Out = os.Stdout
	} else {
		oneLogger.Out = kitsFile.MustOpenLogFile(logCf.LogFile)
	}

	if logCf.LogFormat == "json" {
		oneLogger.Formatter = GetJsonFormat()
	} else {
		oneLogger.Formatter = GetTextFormat()
	}

	if !CheckLogLevel(logCf.LogLevel) {
		fmt.Printf("unsupported loglevel '%s', set it to %s", logCf.LogLevel, WARNING)
		logCf.LogLevel = WARNING
	}

	oneLogger.Level = LogLevelToLogrusLevel[logCf.LogLevel]

	return oneLogger

}

func WriteToLogNoExtraMsg(logWr *myLogger.Logger, fields myLogger.Fields, level string) {
	WriteToLog(logWr, fields, "", level)
}

func WriteToLog(logWr *myLogger.Logger, fields myLogger.Fields, msg string, level string) {
	//fields["errcode"] = errCode

	switch level {
	case ERROR:
		logWr.WithFields(fields).Errorln(msg)
	case WARNING:
		logWr.WithFields(fields).Warningln(msg)
	case INFO:
		logWr.WithFields(fields).Infoln(msg)
	case DEBUG:
		logWr.WithFields(fields).Debugln(msg)
	default:
		logWr.WithFields(fields).Infoln(msg)
	}
}

func WriteLogOnlyMsg(logWr *myLogger.Logger, msg string, level string) {
	switch level {
	case ERROR:
		logWr.WithFields(myLogger.Fields{}).Errorln(msg)

	case WARNING:
		logWr.Warningln(msg)
	case INFO:
		logWr.Infoln(msg)
	case DEBUG:
		logWr.Debugln(msg)
	default:
		logWr.Infoln(msg)
	}
}

type MyLog struct {
	Logger *myLogger.Logger
	LogCfg LogConf
	Wlock  *sync.RWMutex
}

func (this *MyLog) CreateNewRawLogger() {

	this.Logger = &myLogger.Logger{}
	this.Logger.Out = os.Stdout
	this.Logger.Formatter = GetTextFormat()
	this.Logger.Level = LogLevelToLogrusLevel[DEBUG]
	this.Wlock = &sync.RWMutex{}
}

func (this *MyLog) SetLogConf() {
	msg := ""
	if this.LogCfg.LogFile != "" {
		this.Logger.Out = kitsFile.MustOpenLogFile(this.LogCfg.LogFile)
		//msg = fmt.Sprintf("%sset log file to %s\n", this.LogCfg.LogFile)
	} else {
		this.Logger.Out = os.Stdout
		//msg = fmt.Sprintf("%sset log file to stdout\n")
	}

	if this.LogCfg.LogFormat == "json" {
		this.Logger.Formatter = GetJsonFormat()
	} else {
		this.Logger.Formatter = GetTextFormat()
	}

	if this.LogCfg.LogLevel != "" {
		if !CheckLogLevel(this.LogCfg.LogLevel) {
			this.LogCfg.LogLevel = WARNING
			msg = fmt.Sprintf("%s\nunsupported loglevel %s, set it to %s", msg, this.LogCfg.LogLevel, WARNING)
		}
	} else {
		this.LogCfg.LogLevel = WARNING
	}
	this.Logger.Level = LogLevelToLogrusLevel[this.LogCfg.LogLevel]
	if msg != "" {
		this.WriteToLogByFieldsNormal(map[string]interface{}{"errcode": ehand.ERR_ERROR, "errmsg": msg}, WARNING)
	}
}

func (this *MyLog) WriteToLogByFields(msgMap map[string]interface{}, level string, iflogStack bool, exitCode int, ifExitProgram bool) {
	this.Wlock.Lock()
	defer this.Wlock.Unlock()
	var (
		msg  string = ""
		vstr string
		ok   bool
	)

	fields := myLogger.Fields{}
	for k, v := range msgMap {
		if k == ehand.NAME_MSG || k == ehand.NAME_EXT_MSG {
			vstr, ok = v.(string)
			if ok {
				if strings.Contains(vstr, "\n") {
					msg += vstr + "\n"
					continue
				}
			}
		}
		fields[k] = v
	}
	if iflogStack {
		msg += "\n" + string(debug.Stack())
		//fields[ehand.NAME_STACKTRACE] = string(debug.Stack())
	}

	switch level {
	case ERROR:
		this.Logger.WithFields(fields).Errorln("")
	case WARNING:
		this.Logger.WithFields(fields).Warningln("")
	case INFO:
		this.Logger.WithFields(fields).Infoln("")
	case DEBUG:
		this.Logger.WithFields(fields).Debugln("")
	default:
		this.Logger.WithFields(fields).Infoln("")
	}
	if msg != "" {
		fmt.Fprintln(this.Logger.Out, msg)
	}

	if ifExitProgram {
		os.Exit(exitCode)
	}
}

func (this *MyLog) WriteToLogByFieldsNormalOnlyMsg(msg string, level string) {
	this.WriteToLogByFields(map[string]interface{}{ehand.NAME_ERRCODE: ehand.ERR_OK, ehand.NAME_MSG: msg}, level, false, 0, false)
}

func (this *MyLog) WriteToLogByFieldsNormalOnlyMsgExitCode(msg string, level string, exitCode int) {
	this.WriteToLogByFields(map[string]interface{}{ehand.NAME_ERRCODE: ehand.ERR_OK, ehand.NAME_MSG: msg}, level, false, exitCode, false)
}

func (this *MyLog) WriteToLogByFieldsNormalOnlyMsgStack(msg string, level string) {
	this.WriteToLogByFields(map[string]interface{}{ehand.NAME_ERRCODE: ehand.ERR_OK, ehand.NAME_MSG: msg}, level, true, 0, false)
}
func (this *MyLog) WriteToLogByFieldsExitMsgNoErr(msg string, level string, exitCode int) {
	this.WriteToLogByFields(map[string]interface{}{ehand.NAME_ERRCODE: exitCode, ehand.NAME_MSG: msg}, level, false, exitCode, true)
}

func (this *MyLog) WriteToLogByFieldsExitMsgNoErrStack(msg string, level string, exitCode int) {
	this.WriteToLogByFields(map[string]interface{}{ehand.NAME_ERRCODE: exitCode, ehand.NAME_MSG: msg}, level, true, exitCode, true)
}

func (this *MyLog) WriteToLogByFieldsNormal(msgMap map[string]interface{}, level string) {
	this.WriteToLogByFields(msgMap, level, false, 0, false)
}

func (this *MyLog) WriteToLogByFieldsStack(msgMap map[string]interface{}, level string) {
	this.WriteToLogByFields(msgMap, level, true, 0, false)
}

func (this *MyLog) WriteToLogByFieldsExit(msgMap map[string]interface{}, level string, exitCode int) {
	this.WriteToLogByFields(msgMap, level, false, exitCode, true)
}

func (this *MyLog) WriteToLogByFieldsStackExit(msgMap map[string]interface{}, level string, exitCode int) {
	this.WriteToLogByFields(msgMap, level, true, exitCode, true)
}

func (this *MyLog) WriteToLogByFieldsError(err error, extMsg string, level string, ifStack bool, errCode int, ifExit bool) {

	if err == nil {
		return
	}
	msgMap := map[string]interface{}{ehand.NAME_ERRCODE: errCode}
	if extMsg != "" {
		msgMap[ehand.NAME_EXT_MSG] = extMsg
	}
	msgMap[ehand.NAME_MSG] = err.Error()

	this.WriteToLogByFields(msgMap, level, ifStack, errCode, ifExit)
}

func (this *MyLog) WriteToLogByFieldsErrorNormal(err error, level string) {
	this.WriteToLogByFieldsError(err, "", level, false, 0, false)
}

func (this *MyLog) WriteToLogByFieldsErrorStack(err error, level string) {
	this.WriteToLogByFieldsError(err, "", level, true, 0, false)
}

func (this *MyLog) WriteToLogByFieldsErrorExit(err error, level string, errCode int) {
	this.WriteToLogByFieldsError(err, "", level, false, errCode, true)
}

func (this *MyLog) WriteToLogByFieldsErrorStackExit(err error, level string, errCode int) {
	this.WriteToLogByFieldsError(err, "", level, true, errCode, true)
}

func (this *MyLog) WriteToLogByFieldsErrorExtramsg(err error, extMsg string, level string) {
	this.WriteToLogByFieldsError(err, extMsg, level, false, 0, false)
}
func (this *MyLog) WriteToLogByFieldsErrorExtramsgExitCode(err error, extMsg string, level string, exitCode int) {
	this.WriteToLogByFieldsError(err, extMsg, level, false, exitCode, false)
}

func (this *MyLog) WriteToLogByFieldsErrorExtramsgStack(err error, extMsg string, level string) {
	this.WriteToLogByFieldsError(err, extMsg, level, true, 0, false)
}

func (this *MyLog) WriteToLogByFieldsErrorExtramsgExit(err error, extMsg string, level string, errCode int) {
	this.WriteToLogByFieldsError(err, extMsg, level, false, errCode, true)
}

func (this *MyLog) WriteToLogByFieldsErrorExtramsgStackExit(err error, extMsg string, level string, errCode int) {
	this.WriteToLogByFieldsError(err, extMsg, level, true, errCode, true)
}

func (this *MyLog) WriteToLogByMsg(msg string, level string, iflogStack bool, exitCode int, ifExitProgram bool) {
	if iflogStack {
		msg = fmt.Sprintf("%s\n\t%s", msg, string(debug.Stack()))
	}
	switch level {
	case ERROR:
		this.Logger.Errorln(msg)
	case WARNING:
		this.Logger.Warningln(msg)
	case INFO:
		this.Logger.Infoln(msg)
	case DEBUG:
		this.Logger.Debugln(msg)
	default:
		this.Logger.Infoln(msg)
	}
	if ifExitProgram {
		os.Exit(exitCode)
	}
}

func (this *MyLog) WriteToLogByMsgNormal(msg string, level string) {
	this.WriteToLogByMsg(msg, level, false, 0, false)
}

func (this *MyLog) WriteToLogByMsgStack(msg string, level string) {
	this.WriteToLogByMsg(msg, level, true, 0, false)
}

func (this *MyLog) WriteToLogByMsgExit(msg string, level string, exitCode int) {
	this.WriteToLogByMsg(msg, level, false, exitCode, true)
}

func (this *MyLog) WriteToLogByMsgStackExit(msg string, level string, exitCode int) {
	this.WriteToLogByMsg(msg, level, true, exitCode, true)
}

func (this *MyLog) ExitProgram(errcode int, msg string) {
	fmt.Println(msg)
	os.Exit(errcode)
}

func (this *MyLog) GenLogFields(msg string, errCode int) map[string]interface{} {
	return map[string]interface{}{
		ehand.NAME_ERRCODE: errCode,
		ehand.NAME_MSG:     msg,
	}
}
