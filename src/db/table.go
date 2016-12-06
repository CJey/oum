package db

import (
	"fmt"
	"strings"
)

type table interface {
	Name() string
	Latest() (ver uint, cols []string)
	Version(ver uint) (cols []string)
	Upgrade(ver uint, oldT string, newT string) error
}

func createTableSQL(t table) string {
	name := t.Name()
	ver, cols := t.Latest()
	colsstr := "    " + strings.Join(cols, ",\n    ")
	return fmt.Sprintf("create table \"%s\" -- version=%d\n(\n%s\n)", name, ver, colsstr)
}

func createTable(name string, ver uint, cols []string) error {
	colsstr := "    " + strings.Join(cols, ",\n    ")
	st := fmt.Sprintf("create table \"%s\" -- version=%d\n(\n%s\n)", name, ver, colsstr)
	_, err := db.Exec(st)
	return err
}

func ShowCreateTable(names ...string) {
	if len(names) == 0 {
		for name, _ := range ts {
			names = append(names, name)
		}
	}
	fmt.Printf("你可以使用命令行工具sqlite3直接对oum的数据库进行管理\n")
	fmt.Printf("提醒: 所有datetime类型的默认表示均为2016-11-24 16:56:00格式的UTC时间\n")
	for _, name := range names {
		fmt.Printf("\n--------\n\n")
		t, ok := ts[name]
		if ok {
			fmt.Printf("%s\n", createTableSQL(t))
		} else {
			fmt.Printf("[Notice] table %s not exists\n", name)
		}
	}
}
