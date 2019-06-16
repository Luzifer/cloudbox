package sync

import log "github.com/sirupsen/logrus"

func (s *Sync) decideAction(syncState *state, fileName string) error {
	var (
		change = syncState.GetChangeFor(fileName)
		logger = log.WithField("filename", fileName)
	)

	switch {
	case !change.Changed():
		// No changes at all: Get out of here
		logger.Debug("File in sync")
		return nil

	case change.HasAll(ChangeLocalUpdate, ChangeRemoteUpdate):
		// We do have local and remote changes: Check both are now the same or leave this to manual resolve
		logger.Warn("File has local and remote updates, sync not possible")

	case change.HasAll(ChangeLocalAdd, ChangeRemoteAdd):
		// Special case: Both are added, check thet are the same file or break
		logger.Debug("File added locally as well as remotely")
		// TODO: Handle special case

	case change.HasAll(ChangeLocalDelete, ChangeRemoteDelete):
		// Special case: Both vanished, we just need to clean up the sync cache
		logger.Debug("File deleted locally as well as remotely")

		if err := s.deleteDBFileInfo(sideLocal, fileName); err != nil {
			logger.WithError(err).Error("Unable to delete local file info")
			return nil
		}

		if err := s.deleteDBFileInfo(sideRemote, fileName); err != nil {
			logger.WithError(err).Error("Unable to delete remote file info")
			return nil
		}

	case change.Is(ChangeLocalAdd) || change.Is(ChangeLocalUpdate):
		logger.Debug("File added or changed locally, uploading...")
		if err := s.transferFile(s.local, s.remote, sideLocal, sideRemote, fileName); err != nil {
			logger.WithError(err).Error("Unable to upload file")
		}

	case change.Is(ChangeLocalDelete):
		logger.Debug("File deleted locally, removing from remote...")
		if err := s.deleteFile(s.remote, fileName); err != nil {
			logger.WithError(err).Error("Unable to delete file from remote")
		}

	case change.Is(ChangeRemoteAdd) || change.Is(ChangeRemoteUpdate):
		logger.Debug("File added or changed remotely, downloading...")
		if err := s.transferFile(s.remote, s.local, sideRemote, sideLocal, fileName); err != nil {
			logger.WithError(err).Error("Unable to download file")
		}

	case change.Is(ChangeRemoteDelete):
		logger.Debug("File deleted remotely, removing from local...")
		if err := s.deleteFile(s.local, fileName); err != nil {
			logger.WithError(err).Error("Unable to delete file from local")
		}

	default:
		// Unhandled case
		logger.WithField("change", change).Warn("Unhandled change case")
	}

	return nil
}
