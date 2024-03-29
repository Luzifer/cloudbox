package sync

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/Luzifer/cloudbox/providers"
)

const schema = `
CREATE TABLE IF NOT EXISTS local_state (
	relative_name TEXT PRIMARY KEY,
	last_modified DATETIME,
	checksum TEXT,
	size INT
);
CREATE TABLE IF NOT EXISTS remote_state (
	relative_name TEXT PRIMARY KEY,
	last_modified DATETIME,
	checksum TEXT,
	size INT
);
`

func (s *Sync) initSchema() error {
	_, err := s.db.Exec(schema)
	return err
}

func (s *Sync) deleteDBFileInfo(side, relativeName string) error {
	// #nosec G201 - fmt is only used to prefix a table with a constant, no user input
	stmt, err := s.db.Prepare(fmt.Sprintf(`DELETE FROM %s_state WHERE relative_name = ?`, side))
	if err != nil {
		return errors.Wrap(err, "Unable to prepare query")
	}

	_, err = stmt.Exec(relativeName)
	return errors.Wrap(err, "Unable to delete file info")
}

func (s *Sync) setDBFileInfo(side string, info providers.FileInfo) error {
	// #nosec G201 - fmt is only used to prefix a table with a constant, no user input
	stmt, err := s.db.Prepare(fmt.Sprintf(
		`INSERT INTO %s_state VALUES(?, ?, ?, ?) 
			ON CONFLICT(relative_name) DO UPDATE SET 
				last_modified=excluded.last_modified, 
				checksum=excluded.checksum,
				size=excluded.size`, side))
	if err != nil {
		return errors.Wrap(err, "Unable to prepare query")
	}

	_, err = stmt.Exec(info.RelativeName, info.LastModified, info.Checksum, info.Size)
	return errors.Wrap(err, "Unable to upsert file info")
}

func (s *Sync) updateStateFromDatabase(st *state) error {
	for _, table := range []string{sideLocal, sideRemote} {
		// #nosec G201 - fmt is only used to prefix a table with a constant, no user input
		rows, err := s.db.Query(fmt.Sprintf("SELECT * FROM %s_state", table))
		if err != nil {
			return errors.Wrapf(err, "Unable to query table %s", table)
		}
		defer rows.Close()

		for rows.Next() {
			info := providers.FileInfo{}
			if err = rows.Scan(&info.RelativeName, &info.LastModified, &info.Checksum, &info.Size); err != nil {
				return errors.Wrap(err, "Unable to read response")
			}
			st.Set(table, sourceDB, info)
		}
	}

	return nil
}
