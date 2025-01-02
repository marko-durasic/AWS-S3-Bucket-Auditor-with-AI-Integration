package logger

import (
	"fmt"
	"log"
	"os"
)

func InitLogger(logPath string) error {
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	log.SetOutput(logFile)
	return nil
}
