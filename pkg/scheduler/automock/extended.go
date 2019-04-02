package automock

func (_m *Logger) ExpectLoggedOnInfo(msg string, keysAndValues ...interface{}) {
	z := []interface{}{msg}
	for _, kv := range keysAndValues {
		z = append(z, kv)
	}
	_m.On("Info", z...)
}

func (_m *Logger) ExpectLoggedOnError(err error, msg string, keysAndValues ...interface{}) {
	z := []interface{}{err, msg}
	for _, kv := range keysAndValues {
		z = append(z, kv)
	}
	_m.On("Error", z...)
}

func (_m *Logger) ExpectLoggedWithValues(keysAndValues ...interface{}) {
	_m.On("WithValues", keysAndValues...).Return(_m)
}
