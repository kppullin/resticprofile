package status

import (
	"encoding/json"
	"os"
	"time"

	"github.com/spf13/afero"
)

// Status of last schedule profile
type Status struct {
	fs        afero.Fs
	filename  string
	Profile   string         `json:"profile"`
	Backup    *CommandStatus `json:"backup,omitempty"`
	Retention *CommandStatus `json:"retention,omitempty"`
	Check     *CommandStatus `json:"check,omitempty"`
}

// CommandStatus is the last command status
type CommandStatus struct {
	Success bool      `json:"success"`
	Time    time.Time `json:"time"`
	Error   string    `json:"error"`
}

// NewStatus returns a new blank status
func NewStatus(fileName string) *Status {
	return &Status{
		fs:       afero.NewOsFs(),
		filename: fileName,
	}
}

// newAferoStatus returns a new blank status for unit test
func newAferoStatus(fs afero.Fs, fileName string) *Status {
	return &Status{
		fs:       fs,
		filename: fileName,
	}
}

// Load existing status; does not complain if the file does not exists, or is not readable
func (s *Status) Load() *Status {
	// we're not bothered if the status cannot be loaded
	file, err := s.fs.Open(s.filename)
	if err != nil {
		return s
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	_ = decoder.Decode(s)
	return s
}

// Name sets the profile name
func (s *Status) Name(name string) *Status {
	s.Profile = name
	return s
}

// BackupSuccess indicates the last backup was successful
func (s *Status) BackupSuccess() *Status {
	s.Backup = newSuccess()
	return s
}

// BackupError sets the error of the last backup
func (s *Status) BackupError(err error) *Status {
	s.Backup = newError(err)
	return s
}

// RetentionSuccess indicates the last retention was successful
func (s *Status) RetentionSuccess() *Status {
	s.Retention = newSuccess()
	return s
}

// RetentionError sets the error of the last retention
func (s *Status) RetentionError(err error) *Status {
	s.Retention = newError(err)
	return s
}

// CheckSuccess indicates the last check was successful
func (s *Status) CheckSuccess() *Status {
	s.Check = newSuccess()
	return s
}

// CheckError sets the error of the last check
func (s *Status) CheckError(err error) *Status {
	s.Check = newError(err)
	return s
}

// Save current status to the file
func (s *Status) Save() error {
	file, err := s.fs.OpenFile(s.filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	err = encoder.Encode(s)
	if err != nil {
		return err
	}
	return nil
}
func newSuccess() *CommandStatus {
	return &CommandStatus{
		Success: true,
		Time:    time.Now(),
	}
}

func newError(err error) *CommandStatus {
	return &CommandStatus{
		Success: false,
		Time:    time.Now(),
		Error:   err.Error(),
	}
}
