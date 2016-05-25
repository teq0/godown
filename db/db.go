package db

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"os"
)

const dbPath = "./db/godown.db"
const dbSchemaVersion = "1.0.0"

func Init() (OK bool, err error) {
	// this function will check the db exists, and if not, create it
	// and check that the schema is up to date and if not, update it

	OK, err = checkDBExists()

	if err == nil {
		if !OK {
			OK, err = createDB()

			if err != nil {
				log.Fatal("DB:Init: error creating new database.", err)
				OK = false
				return
			}
		}
	} else {
		log.Fatal("DB::Init: error checking database exists", err)
		panic(err)
	}

	OK, err = checkDBSchemaVersion()

	if err != nil {
		log.Fatal("DB:Init: error reading database schema version.", err)
		OK = false
		return
	}

	if err == nil {
		if !OK {
			OK, err = updateDBSchema()
		}
	} else {
		log.Fatal("DB:Init: error updating database schema.", err)
	}

	return OK, err
}

func checkDBExists() (dbExists bool, err error) {
	dbExists = false

	db, err := sql.Open("sqlite3", dbPath)
	defer db.Close()

	if err == nil {
		rows, err := db.Query("select * from dbinfo")

		if err == nil {
			if rows.Next() {

				dbExists = true
			}
		}
	}

	return
}

func createDB() (OK bool, err error) {
	copyFileContents("./db/empty.db", dbPath)
	db, err := sql.Open("sqlite3", dbPath)
	defer db.Close()

	return
}

func checkDBSchemaVersion() (schemaOK bool, err error) {
	db, err := sql.Open("sqlite3", dbPath)
	defer db.Close()

	var schemaVer string

	if err == nil {
		rows, err := db.Query("select top 1 schema_version from dbinfo")
		if err == nil {
			for rows.Next() {
				rows.Scan(&schemaVer)

				if schemaVer == dbSchemaVersion {
					schemaOK = true
				}
			}
		}
	}

	return
}

func updateDBSchema() (OK bool, err error) {
	OK = true

	return
}

// TODO - find a library that does this

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
