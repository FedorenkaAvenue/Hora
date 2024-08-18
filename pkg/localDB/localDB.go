// Stores DB as json file.

package localdb

import (
	"encoding/json"
	"os"
)

type LocalDB struct {
	filePath string
	Data     db
}

type db map[string]DBItems
type DBItems map[string]string

func (d LocalDB) New(filePath string) LocalDB {
	var l LocalDB
	l.filePath = filePath

	file, err := os.ReadFile(l.filePath)

	if err != nil {
		err := os.WriteFile(l.filePath, []byte("{}"), 0755)

		if err != nil {
			panic(err)
		}

		l.Data = make(db)
	} else {
		err := json.Unmarshal(file, &l.Data)

		if err != nil {
			panic(err)
		}
	}

	return l
}

func (d *LocalDB) Update(k string, v DBItems) error {
	dbCopy := make(db)

	for k, v := range d.Data {
		dbCopy[k] = v
	}

	dbCopy[k] = v

	j, jErr := json.MarshalIndent(dbCopy, "", "\t")

	if jErr != nil {
		return jErr
	}

	tErr := os.WriteFile(d.filePath, j, 0755)

	if tErr != nil {
		return tErr
	}

	d.Data = dbCopy

	return nil
}
