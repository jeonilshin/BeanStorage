package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
)

var tpl *template.Template
var db *sql.DB

func main() {
	tpl, _ = template.ParseGlob("templates/*.html")
	var err error
	db, err = sql.Open("mysql", "root:password@tcp(localhost:3306)/beanstorage")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	http.HandleFunc("/register", register)
	http.HandleFunc("/registerauth", registerAuthHandler)
	http.ListenAndServe("localhost:8080", nil)
}