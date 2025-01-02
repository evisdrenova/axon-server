package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Creates a file to log the logs to
func CreateFileLogger() (*log.Logger, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("error getting executable path: %v", err)
	}

	logsDir := filepath.Join(filepath.Dir(execPath), "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating logs directory: %v", err)
	}

	logFileName := filepath.Join(logsDir, fmt.Sprintf("api_requests_%s.log", time.Now().Format("2006-01-02")))
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("error opening log file: %v", err)
	}

	log.Printf("Logging to: %s", logFileName)
	return log.New(logFile, "", log.LstdFlags), nil
}
