// Stores DB as json file.

package localdb

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type LocalDB struct {
	Data           data
	ItemLifePeriod int64
	StorePath      string
}

type data map[string]db
type db map[string]int64

func (l *LocalDB) Init(dbNames []string) {
	l.Data = make(data)

	if _, err := os.Stat(l.StorePath); os.IsNotExist(err) {
		err := os.Mkdir(l.StorePath, 0755)

		if err != nil {
			panic(err)
		}
	}

	for _, d := range dbNames {
		l.Data[d] = make(db)
		l.initDB(d)
	}

	go l.cronClear()
}

func (l *LocalDB) Append(dbName string, items []string) error {
	now := time.Now()
	formattedDate := now.Unix()

	for _, i := range items {
		l.Data[dbName][i] = formattedDate
	}

	err := l.refreshDBFile(dbName)

	if err != nil {
		return err
	}

	return nil
}

func (l *LocalDB) initDB(name string) {
	fileName := fmt.Sprintf("%v/%v.json", l.StorePath, name)
	file, err := os.ReadFile(fileName)

	if err != nil {
		err := os.WriteFile(fileName, []byte("[]"), 0755)

		if err != nil {
			panic(err)
		}
	} else {
		t := l.Data[name]
		err := json.Unmarshal(file, &t)

		if err != nil {
			panic(err)
		}
	}
}

func (l *LocalDB) refreshDBFile(dbName string) error {
	j, jErr := json.MarshalIndent(l.Data[dbName], "", "\t")

	if jErr != nil {
		return jErr
	}

	tErr := os.WriteFile(fmt.Sprintf("%v/%v.json", l.StorePath, dbName), j, 0755)

	if tErr != nil {
		return tErr
	}

	return nil
}

func (l *LocalDB) cronClear() {
	for {
		now := time.Now().Unix()

		for dbName, db := range l.Data {
			for k := range db {
				if now-db[k] > l.ItemLifePeriod {
					delete(db, k)
				}
			}

			l.refreshDBFile(dbName)
		}

		<-time.After(time.Hour * 24)
	}
}
