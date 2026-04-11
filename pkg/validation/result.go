package validation

type Level int

const (
	LevelWarning Level = iota
	LevelError
)

type ValidationResult struct {
	Ok bool
	Level Level
	Message string
}
