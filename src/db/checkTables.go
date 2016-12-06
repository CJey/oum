package db

import (
	"database/sql"
	"fmt"
	"math/rand"
	//"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/cjey/slog"
)

var ts map[string]table = map[string]table{
	tUser{}.Name():     tUser{},
	tDevice{}.Name():   tDevice{},
	tIfconfig{}.Name(): tIfconfig{},
	tActive{}.Name():   tActive{},
	tOVPN{}.Name():     tOVPN{},
	tLog{}.Name():      tLog{},
	tIPCache{}.Name():  tIPCache{},
}

func checkTables() (err error) {
	for _, t := range ts {
		err = checkTable(t)
		if err != nil {
			return
		}
	}
	return
}

func checkTable(t table) error {
	name := t.Name()
	ver, cols := t.Latest()

	var schema string
	err := db.QueryRow("select sql from sqlite_master where type='table' and name=?", name).
		Scan(&schema)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == sql.ErrNoRows {
		// new table
		slog.Infof("create new table %s ver %d", name, ver)
		return createTable(name, ver, cols)
	}

	// table exists
	idx := strings.Index(schema, "\n")
	if idx < 0 {
		return fmt.Errorf("Invalid table[%s] schema", name)
	}
	firstline := schema[0:idx]
	tmp := strings.SplitN(firstline, "--", 2)
	if len(tmp) < 2 {
		return fmt.Errorf("Invalid table[%s] schema", name)
	}
	comment := strings.TrimSpace(tmp[1])
	cver := parseVersion(comment)
	if cver == 0 {
		return fmt.Errorf("Invalid table[%s] schema", name)
	}
	if cver == ver {
		return nil
	}
	if cver > ver {
		return fmt.Errorf("Table[%s] version too high", name)
	}

	oldT := name
	newT := name
	// upgrade table
	slog.Infof("upgrade table %s %d => %d", name, cver, ver)
	for cver < ver {
		cver++
		newT = fmt.Sprintf("%s_%d", name, rand.Int63())
		err = createTable(newT, cver, t.Version(cver))
		if err != nil {
			return err
		}
		err = t.Upgrade(cver, oldT, newT)
		if err != nil {
			return err
		}
		oldT = newT
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("drop table " + name)
	if err != nil {
		return err
	}
	_, err = tx.Exec("alter table " + newT + " rename to " + name)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func parseVersion(comment string) uint {
	matches := regexp.MustCompile(`\bversion=(\d+)\b`).FindStringSubmatch(comment)
	if len(matches) == 2 {
		v, _ := strconv.Atoi(matches[1])
		return uint(v)
	}
	return 0
}
