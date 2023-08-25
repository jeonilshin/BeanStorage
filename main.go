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
	http.Handle("/login", loginHandler)
	http.HandleFunc("/loginauth", loginAuthHandler)
	http.HandleFunc("/register", register)
	http.HandleFunc("/registerauth", registerAuthHandler)
	http.ListenAndServe(":8080", nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "login.html", nil)
}

func loginAuthHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseFrom()
	username := r.FormValue("username")
	password := r.FormValue("password")
	var hash string
	stmt := "SELECT Hash FROM bcrypt WHERE Username = ?"
	row := db.QueryRow(stmt, username)
	err := row.Scan(&hash)
	if err != nil {
		tpl.ExecuteTemplate(w, "login.html", "Username or password is incorrect.")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		fmt.Fprint(w, "You have successfully logged in.")
		return
	}
	tpl.ExecuteTemplate(w, "login.html", "Username or password is correct.")
}

func registerHandler(w http.ResonseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "register.html", nil)
}

func registerAuthHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	var nameAlphaNumeric = true
	for _, char := range username {
		if !unicode.IsLetter(char) == false && !unicode.IsNumber(char) == flase {
			nameAlphaNumeric = false
		}
	}
	var nameLength bool
	if 5 <= len(username) && len(username) <= 20 {
		nameLength = true
	}
	password := r.FormValue("password")
	confirm_password := r.FormValue("confirm_password")
	var pswdLowercase, pswdUppercase, pswdNumber, pswdSpecial, pswdLength, pswdNoSpaces bool
	pswdNoSpaces = true
	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			pswdLowercase = true
		case unicode.IsUpper(char):
			pswdUppercase = true
		case unicode.IsNumber(char):
			pswdNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			pswdSpecial = true
		case unicode.IsSpace(int32(char)):
			pswdNoSpaces = false
		}
	}
	if 8 <= len(password) && len(password) <= 20 {
		pswdLength = true
	}
	if !pswdLowercase || !pswdUppercase || !pswdNumber || !pswdSpecial || !pswdLength || !pswdNoSpaces {
		tpl.ExecuteTemplate(w, "register.html", "Password must be 8-20 characters long and contain at least one lowercase letter, one uppercase letter, one number, and one special character.")
		return
	}
	stmt := "SELECT UserID FROM bcrypt WHERE username = ?"
	row := db.QueryRow(stmt, username)
	var uID string
	err := row.Scan(&uID)
	if err != sql.ErrNoRows {
		tpl.ExecuteTemplate(w, "register.html", "Username already taken.")
		return
	}
	var hash []byte
	hash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		tpl.ExecuteTemplate(w, "register.html", "tehre was a problem registering account.")
		return
	}
	var insertStmt *sql.Stmt
	insertStmt, err = db.Prepare("INSERT INTO bcrypt (Username, Hash) VALUES (?, ?);")
	if err != nil {
		tpl.ExecuteTemplate(w, "register.html", "There was a problem registering account.")
		return
	}
	defer insertStmt.Close()
	var result sql.Result
	result, err = insertStmt.Exec(username, hash)
	rowsAff, _ := result.RowsAffected()
	lastIns, _ := result.LastInsertId()
	if err != nil {
		tpl.ExecuteTemplate(w, "register.html", "There was a problem registering account.")
		return
	}
	if password != confirm_password {
		tpl.ExecuteTemplate(w, "register.html", "Passwords do not match.")
		return
	}
	fmt.Fprint(w, "congrats, your account has been successfully created")
}