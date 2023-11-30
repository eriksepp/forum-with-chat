package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"forum/application"
	"forum/model/sqlpkg"
	"forum/route"
)

const DB_Name = "forumDB.db"

func main() {
	// app keeps all dependences used by handlers
	app, err := application.New()
	if err != nil {
		app.ErrLog.Fatalln(err)
	}

	port, pristineDB, testDB, err := parseArgs()
	if err != nil {
		app.ErrLog.Fatalln(err)
	}

	// init DB pool DB_Name
	_, err = os.Stat(DB_Name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			createAndFillTestDB(app)
		} else {
			app.ErrLog.Fatalln(err)
		}
	} else {
		switch {
		case testDB:
			if os.Rename(DB_Name, DB_Name+".bak") != nil {
				app.ErrLog.Fatalln("cannot rename the DB file")
			}
			createAndFillTestDB(app)
		case pristineDB:
			// rename DB file
			if os.Rename(DB_Name, DB_Name+".bak") != nil {
				app.ErrLog.Fatalln("cannot rename the DB file")
			}

			err := app.CreateDB(DB_Name)
			if err != nil {
				app.ErrLog.Fatalln(err)
			}
		default:
			db, err := sqlpkg.OpenDB(DB_Name, "webuser", "webuser")
			if err != nil {
				app.ErrLog.Fatalln(err)
			}
			app.ForumData = &sqlpkg.ForumModel{DB: db}
		}
	}
	defer app.ForumData.DB.Close()

	// Starting the web server
	server := &http.Server{
		Addr:     fmt.Sprintf("localhost:%s", port),
		ErrorLog: app.ErrLog,
		Handler:  route.Load(app),
	}
	app.Server= server
	fmt.Printf("Starting server at http://www.localhost:%s\n\n", port)
	app.InfoLog.Printf("Starting server at port %s\n", port)
	if err := server.ListenAndServe(); err != nil {
		app.ErrLog.Fatal(err)
	}
}

// Parses the program's arguments to obtain the server port. If no arguments found, it uses the 8080 port by default
// Usage: go run .  --port=PORT_NUMBER --pristine --testdb
func parseArgs() (port string, pristineDB bool, testDB bool, err error) {
	usage := `wrong arguments
     Usage: go run ./app [OPTIONS]
     OPTIONS: 
            --port=PORT_NUMBER
            --p=PORT_NUMBER
            --pristine to drop the existing DB and create the new one from scratch
            --testdb drop the existing DB and start with the test DB`
	flag.StringVar(&port, "port", "8080", "server port")
	flag.StringVar(&port, "p", "8080", "server port (shorthand)")
	flag.BoolVar(&pristineDB, "pristine", false, "--pristine if you want drop the existing DB and create the new one from scratch")
	flag.BoolVar(&testDB, "testdb", false, "--testdb if you want drop the existing DB and start with the test DB")
	flag.Parse()
	if flag.NArg() > 0 {
		return "", false, false, fmt.Errorf(usage)
	}
	_, err = strconv.ParseUint(port, 10, 16)
	if err != nil {
		return "", false, false, fmt.Errorf("error: port must be a 16-bit unsigned number ")
	}
	return
}

func createAndFillTestDB(app *application.Application) {
	err := app.CreateDB(DB_Name)
	if err != nil {
		app.ErrLog.Fatalln(err)
	}
	err = app.FillTestDB("model/sqlpkg/testDB.sql")
	if err != nil {
		app.ErrLog.Fatalln(err)
	}
}
