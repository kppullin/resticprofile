package status

import (
	"errors"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestLoadNoFile(t *testing.T) {
	name := "test profile"
	status := NewStatus("some file").Name(name)
	assert.Equal(t, name, status.Profile)
	status.Load()
	// Profile name should not have been cleared
	assert.Equal(t, name, status.Profile)
}

func TestBackupSuccess(t *testing.T) {
	status := NewStatus("")
	assert.Nil(t, status.Backup)
	status.BackupSuccess()
	assert.True(t, status.Backup.Success)
	assert.Empty(t, status.Backup.Error)
}

func TestBackupError(t *testing.T) {
	errorMessage := "test test test"
	status := NewStatus("")
	assert.Nil(t, status.Backup)
	status.BackupError(errors.New(errorMessage))
	assert.False(t, status.Backup.Success)
	assert.Equal(t, errorMessage, status.Backup.Error)
}

func TestRetentionSuccess(t *testing.T) {
	status := NewStatus("")
	assert.Nil(t, status.Retention)
	status.RetentionSuccess()
	assert.True(t, status.Retention.Success)
	assert.Empty(t, status.Retention.Error)
}

func TestRetentionError(t *testing.T) {
	errorMessage := "test test test"
	status := NewStatus("")
	assert.Nil(t, status.Retention)
	status.RetentionError(errors.New(errorMessage))
	assert.False(t, status.Retention.Success)
	assert.Equal(t, errorMessage, status.Retention.Error)
}

func TestCheckSuccess(t *testing.T) {
	status := NewStatus("")
	assert.Nil(t, status.Check)
	status.CheckSuccess()
	assert.True(t, status.Check.Success)
	assert.Empty(t, status.Check.Error)
}

func TestCheckError(t *testing.T) {
	errorMessage := "test test test"
	status := NewStatus("")
	assert.Nil(t, status.Check)
	status.CheckError(errors.New(errorMessage))
	assert.False(t, status.Check.Success)
	assert.Equal(t, errorMessage, status.Check.Error)
}

func TestSaveAndLoadEmptyStatus(t *testing.T) {
	filename := "TestSaveAndLoadEmptyStatus.json"

	fs := afero.NewMemMapFs()
	status := newAferoStatus(fs, filename)
	err := status.Save()
	assert.NoError(t, err)

	exists, err := afero.Exists(fs, filename)
	assert.NoError(t, err)
	assert.True(t, exists)

	status = newAferoStatus(fs, filename).Load()
	assert.Equal(t, "", status.Profile)
	assert.Nil(t, status.Backup)
	assert.Nil(t, status.Retention)
	assert.Nil(t, status.Check)
}

func TestSaveAndLoadBackupSuccess(t *testing.T) {
	filename := "TestSaveAndLoadBackupSuccess.json"

	fs := afero.NewMemMapFs()
	status := newAferoStatus(fs, filename).Load().BackupSuccess()
	err := status.Save()
	assert.NoError(t, err)

	status = newAferoStatus(fs, filename).Load()
	assert.NotNil(t, status.Backup)
	assert.Nil(t, status.Retention)
	assert.Nil(t, status.Check)
	assert.True(t, status.Backup.Success)
}

func TestSaveAndLoadBackupError(t *testing.T) {
	filename := "TestSaveAndLoadBackupError.json"
	errorMessage := "message in a box"

	fs := afero.NewMemMapFs()
	status := newAferoStatus(fs, filename).Load().BackupError(errors.New(errorMessage))
	err := status.Save()
	assert.NoError(t, err)

	status = newAferoStatus(fs, filename).Load()
	assert.NotNil(t, status.Backup)
	assert.Nil(t, status.Retention)
	assert.Nil(t, status.Check)
	assert.False(t, status.Backup.Success)
	assert.Equal(t, errorMessage, status.Backup.Error)
}
