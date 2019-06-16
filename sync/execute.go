package sync

import (
	"github.com/pkg/errors"

	"github.com/Luzifer/cloudbox/providers"
)

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
