package logger

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	WithFields(fields map[string]any) Logger
}

type Field struct {
	Key   string
	Value any
}

func String(key string, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Err(err error) Field {
	return Field{
		Key:   "error",
		Value: err.Error(),
	}
}
