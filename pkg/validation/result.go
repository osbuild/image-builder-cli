package validation

type Level int

const (
	LevelWarning Level = iota
	LevelError
)

type ValidationResult struct {
	Level Level
	Message string
}
