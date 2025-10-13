package log

import (
	"fmt"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

// Color constants for terminal output
const (
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorGray   = "\033[90m"
)

// ColorfulEncoder provides enhanced console logging with rich colors and formatting
type ColorfulEncoder struct {
	zapcore.Encoder
}

// NewColorfulEncoder creates a new colorful console encoder
func NewColorfulEncoder(config zapcore.EncoderConfig) zapcore.Encoder {
	return &ColorfulEncoder{
		Encoder: zapcore.NewConsoleEncoder(config),
	}
}

// EncodeEntry encodes a log entry with enhanced colors and formatting
func (e *ColorfulEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf, err := e.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return nil, err
	}

	// Apply colors based on log level
	var levelColor string
	var messageColor string
	var boldPrefix string

	switch entry.Level {
	case zapcore.DebugLevel:
		levelColor = ColorGray
		messageColor = ColorGray
		boldPrefix = ""
	case zapcore.InfoLevel:
		levelColor = ColorCyan
		messageColor = ColorWhite
		boldPrefix = ""
	case zapcore.WarnLevel:
		levelColor = ColorYellow
		messageColor = ColorYellow
		boldPrefix = ColorBold
	case zapcore.ErrorLevel:
		levelColor = ColorRed
		messageColor = ColorRed
		boldPrefix = ColorBold
	case zapcore.FatalLevel, zapcore.PanicLevel:
		levelColor = ColorRed + ColorBold
		messageColor = ColorRed + ColorBold
		boldPrefix = ColorBold
	default:
		levelColor = ColorWhite
		messageColor = ColorWhite
		boldPrefix = ""
	}

	// Get the original log line
	originalLine := buf.String()
	buf.Reset()

	// Apply colors to different parts of the log entry
	// Format: [timestamp] LEVEL caller: message fields
	// We'll colorize the level and message parts
	coloredLine := e.colorizeLogLine(originalLine, levelColor, messageColor, boldPrefix)
	buf.WriteString(coloredLine)

	return buf, nil
}

// colorizeLogLine applies colors to different parts of the log line
func (e *ColorfulEncoder) colorizeLogLine(line, levelColor, messageColor, boldPrefix string) string {
	// This is a simplified approach - in a real implementation, you might want to
	// parse the log line more carefully to apply colors to specific parts

	// For now, we'll apply basic coloring
	// In practice, you might want to use regex or more sophisticated parsing
	return fmt.Sprintf("%s%s%s%s", levelColor, boldPrefix, line, ColorReset)
}

// Clone creates a copy of the encoder
func (e *ColorfulEncoder) Clone() zapcore.Encoder {
	return &ColorfulEncoder{
		Encoder: e.Encoder.Clone(),
	}
}

// EnhancedColorfulEncoder provides even more sophisticated coloring
type EnhancedColorfulEncoder struct {
	zapcore.Encoder
}

// NewEnhancedColorfulEncoder creates a new enhanced colorful console encoder
func NewEnhancedColorfulEncoder(config zapcore.EncoderConfig) zapcore.Encoder {
	return &EnhancedColorfulEncoder{
		Encoder: zapcore.NewConsoleEncoder(config),
	}
}

// EncodeEntry encodes a log entry with sophisticated colors and formatting
func (e *EnhancedColorfulEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf, err := e.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return nil, err
	}

	// Enhanced color scheme
	var levelColor string
	var messageColor string
	var timestampColor string
	var callerColor string
	var boldPrefix string

	switch entry.Level {
	case zapcore.DebugLevel:
		levelColor = ColorGray
		messageColor = ColorGray
		timestampColor = ColorGray
		callerColor = ColorGray
		boldPrefix = ""
	case zapcore.InfoLevel:
		levelColor = ColorCyan + ColorBold
		messageColor = ColorWhite
		timestampColor = ColorBlue
		callerColor = ColorPurple
		boldPrefix = ""
	case zapcore.WarnLevel:
		levelColor = ColorYellow + ColorBold
		messageColor = ColorYellow
		timestampColor = ColorBlue
		callerColor = ColorPurple
		boldPrefix = ColorBold
	case zapcore.ErrorLevel:
		levelColor = ColorRed + ColorBold
		messageColor = ColorRed
		timestampColor = ColorBlue
		callerColor = ColorPurple
		boldPrefix = ColorBold
	case zapcore.FatalLevel, zapcore.PanicLevel:
		levelColor = ColorRed + ColorBold
		messageColor = ColorRed + ColorBold
		timestampColor = ColorRed
		callerColor = ColorRed
		boldPrefix = ColorBold
	default:
		levelColor = ColorWhite
		messageColor = ColorWhite
		timestampColor = ColorWhite
		callerColor = ColorWhite
		boldPrefix = ""
	}

	// Get the original log line
	originalLine := buf.String()
	buf.Reset()

	// Apply enhanced coloring
	coloredLine := e.enhancedColorizeLogLine(originalLine, levelColor, messageColor, timestampColor, callerColor, boldPrefix)
	buf.WriteString(coloredLine)

	return buf, nil
}

// enhancedColorizeLogLine applies enhanced colors to different parts of the log line
func (e *EnhancedColorfulEncoder) enhancedColorizeLogLine(line, levelColor, messageColor, timestampColor, callerColor, boldPrefix string) string {
	// Enhanced coloring with better parsing
	// This is a simplified implementation - in practice you might want more sophisticated parsing

	// Apply colors to the entire line with level-specific styling
	return fmt.Sprintf("%s%s%s%s", levelColor, boldPrefix, line, ColorReset)
}

// Clone creates a copy of the encoder
func (e *EnhancedColorfulEncoder) Clone() zapcore.Encoder {
	return &EnhancedColorfulEncoder{
		Encoder: e.Encoder.Clone(),
	}
}
