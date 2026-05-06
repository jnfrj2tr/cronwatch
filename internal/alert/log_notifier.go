package alert

import "log"

// LogNotifier writes alerts to the standard logger.
type LogNotifier struct{}

// NewLogNotifier returns a LogNotifier.
func NewLogNotifier() *LogNotifier {
	return &LogNotifier{}
}

// Send logs the alert using the standard log package.
func (l *LogNotifier) Send(a Alert) error {
	log.Println(FormatMessage(a))
	return nil
}
