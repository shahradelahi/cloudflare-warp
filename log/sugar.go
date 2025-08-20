package log

func logf(lvl Level, template string, args ...interface{}) {
	_globalMu.RLock()
	s := _globalS
	_globalMu.RUnlock()
	s.Logf(lvl, template, args...)
}

func Debugf(template string, args ...interface{}) {
	logf(DebugLevel, template, args...)
}

func Infof(template string, args ...interface{}) {
	logf(InfoLevel, template, args...)
}

func Warnf(template string, args ...interface{}) {
	logf(WarnLevel, template, args...)
}

func Errorf(template string, args ...interface{}) {
	logf(ErrorLevel, template, args...)
}

func Fatalf(template string, args ...interface{}) {
	logf(FatalLevel, template, args...)
}

func logw(lvl Level, msg string, keysAndValues ...interface{}) {
	_globalMu.RLock()
	s := _globalS
	_globalMu.RUnlock()
	s.Logw(lvl, msg, keysAndValues...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	logw(DebugLevel, msg, keysAndValues...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	logw(InfoLevel, msg, keysAndValues...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	logw(WarnLevel, msg, keysAndValues...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	logw(ErrorLevel, msg, keysAndValues...)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	logw(FatalLevel, msg, keysAndValues...)
}
