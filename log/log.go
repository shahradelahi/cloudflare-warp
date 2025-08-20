package log

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// global Logger and SugaredLogger.
var (
	_globalMu sync.RWMutex
	_globalL  *Logger
	_globalS  *SugaredLogger
)

func init() {
	SetLogger(zap.Must(zap.NewProduction()))
}

func NewLeveled(l Level, options ...Option) (*Logger, error) {
	switch l {
	case SilentLevel:
		return zap.NewNop(), nil
	case DebugLevel:
		return zap.NewDevelopment(options...)
	case InfoLevel, WarnLevel, ErrorLevel, DPanicLevel, PanicLevel, FatalLevel:
		cfg := zap.NewProductionConfig()
		cfg.Level.SetLevel(l)
		return cfg.Build(options...)
	default:
		return nil, fmt.Errorf("invalid level: %s", l)
	}
}

// SetLogger sets the global Logger and SugaredLogger.
func SetLogger(logger *Logger) {
	_globalMu.Lock()
	defer _globalMu.Unlock()
	// apply pkgCallerSkip to global loggers.
	_globalL = logger.WithOptions(pkgCallerSkip)
	_globalS = _globalL.Sugar()
}

// GetLogger returns the global logger.
func GetLogger() *Logger {
	_globalMu.RLock()
	defer _globalMu.RUnlock()
	return _globalL
}

func log(lvl Level, args ...interface{}) {
	_globalMu.RLock()
	s := _globalS
	_globalMu.RUnlock()
	s.Log(lvl, args...)
}

func Debug(args ...interface{}) {
	log(DebugLevel, args...)
}

func Info(args ...interface{}) {
	log(InfoLevel, args...)
}

func Warn(args ...interface{}) {
	log(WarnLevel, args...)
}

func Error(args ...interface{}) {
	log(ErrorLevel, args...)
}

func Fatal(args ...interface{}) {
	log(FatalLevel, args...)
}
