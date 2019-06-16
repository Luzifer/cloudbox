package sync

import (
	"crypto/sha256"

	"github.com/pkg/errors"

	"github.com/Luzifer/cloudbox/providers"
)

func (s *Sync) addBothCreated(fileName string) error {
	// Use forced sha256 to ensure lesser chance for collision
	var hashMethod = sha256.New()

	local, err := s.local.GetFile(fileName)
	if err != nil {
		return errors.Wrap(err, "Unable to retrieve file from local")
	}

	remote, err := s.remote.GetFile(fileName)
	if err != nil {
		return errors.Wrap(err, "Unable to retrieve file from remote")
	}

	localSum, err := local.Checksum(hashMethod)
	if err != nil {
		return errors.Wrap(err, "Unable to get checksum from local file")
	}

	remoteSum, err := remote.Checksum(hashMethod)
	if err != nil {
		return errors.Wrap(err, "Unable to get checksum from remote file")
	}

	if localSum != remoteSum {
		return errors.New("Checksums differ")
	}

	if err := s.setDBFileInfo(sideLocal, local.Info()); err != nil {
		return errors.Wrap(err, "Unable to update DB info for local file")
	}

	if err := s.setDBFileInfo(sideRemote, remote.Info()); err != nil {
		return errors.Wrap(err, "Unable to update DB info for remote file")
	}

	return nil
}

func (s *Sync) deleteFile(on providers.CloudProvider, fileName string) error {
	if err := on.DeleteFile(fileName); err != nil {
		return errors.Wrap(err, "Unable to delete file")
	}

	if err := s.deleteDBFileInfo(sideLocal, fileName); err != nil {
		return errors.Wrap(err, "Umable to delete local file info")
	}

	if err := s.deleteDBFileInfo(sideRemote, fileName); err != nil {
		return errors.Wrap(err, "Umable to delete remote file info")
	}

	return nil
}

func (s *Sync) transferFile(from, to providers.CloudProvider, sideFrom, sideTo, fileName string) error {
	file, err := from.GetFile(fileName)
	if err != nil {
		return errors.Wrap(err, "Unable to retrieve file")
	}

	newFile, err := to.PutFile(file)
	if err != nil {
		return errors.Wrap(err, "Unable to put file")
	}

	if err := s.setDBFileInfo(sideTo, newFile.Info()); err != nil {
		return errors.Wrap(err, "Unable to update DB info for target file")
	}

	if err := s.setDBFileInfo(sideFrom, file.Info()); err != nil {
		return errors.Wrap(err, "Unable to update DB info for source file")
	}

	return nil
}
