package mylog

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	myLog      *log.Logger
	logFile    *os.File
	filePrefix string
	showFile   bool
)

// Init initializes the logger with a log file name prefix and an option to show file name and line number
func Init(logFileNamePrefix string, showFilename bool) error {
	filePrefix = logFileNamePrefix
	showFile = showFilename

	if showFile {
		myLog = log.New(logFile, "", log.LstdFlags|log.Llongfile)
	} else {
		myLog = log.New(logFile, "", log.LstdFlags)
	}

	return nil
}

// getLogFileName generates the log file name based on the current week
func getLogFileName() string {
	now := time.Now()
	year, week := now.ISOWeek()
	return fmt.Sprintf("%s_%d_W%d.log", filePrefix, year, week)
}

// rotateLogFile creates or opens the log file for the current week
func rotateLogFile() error {
	logFileName := getLogFileName()

	// Close the old log file if it is open
	if logFile != nil {
		logFile.Close()
	}

	var err error
	logFile, err = os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	myLog.SetOutput(logFile)
	return nil
}

// Debug logs a debug-level message with format
func Debug(format string, args ...interface{}) {
	logWithPrefix("[DEBUG]", format, args...)
}

// Info logs an info-level message with format
func Info(format string, args ...interface{}) {
	logWithPrefix("[INFO]", format, args...)
}

// Warn logs a warn-level message with format
func Warn(format string, args ...interface{}) {
	logWithPrefix("[WARN]", format, args...)
}

// Error logs an error-level message with format
func Error(format string, args ...interface{}) {
	logWithPrefix("[ERROR]", format, args...)
}

// logWithPrefix handles the log output with a specific prefix and supports formatting
func logWithPrefix(prefix, format string, args ...interface{}) {
	now := time.Now()
	message := fmt.Sprintf(format, args...)
	fullMessage := fmt.Sprintf("%s %s %s", prefix, now.Format("2006-01-02 15:04:05"), message)

	// Check if the log file needs to be rotated
	expectedFileName := getLogFileName()
	if logFile == nil || (logFile.Name() != expectedFileName) {
		if err := rotateLogFile(); err != nil {
			fmt.Println("Error rotating log file:", err)
			return
		}
	}

	_ = myLog.Output(3, fullMessage)
}

// Close safely closes the log file
func Close() {
	if logFile != nil {
		if err := logFile.Close(); err != nil {
			fmt.Println("Error closing log file:", err)
		}
	}
}
