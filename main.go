package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"unicode"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
)

var tpl *template.Template
var db *sql.DB

var store = sessions.NewCookieStore([]byte("super-secret"))

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
    http.HandleFunc("/logout", logoutHandler)
    http.HandleFunc("/register", registerHandler)
    http.HandleFunc("/registerauth", registerAuthHandler)
    http.HandleFunc("/about", Auth(aboutHandler))
    http.HandleFunc("/", Auth(indexHandler))
    http.ListenAndServe(":8080", context.ClearHandler(http.DefaultServeMux))
}

func Auth(HandlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		_, ok := session.Values["userID"]
		if !ok {
			http.Redirect(w, r, "/login", 302) // StatusFound = 302
			return
		}
		HandlerFunc(w, r)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "login.html", nil)
}

func loginAuthHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    username := r.FormValue("username")
    password := r.FormValue("password")
    var hash, userID string
    stmt := "SELECT Hash, UserID FROM bcrypt WHERE Username = ?"
    row := db.QueryRow(stmt, username)
    err := row.Scan(&hash, &userID)
    if err != nil {
        tpl.ExecuteTemplate(w, "login.html", "Username or password is incorrect.")
        return
    }

    err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    if err != nil {
        session, _ := store.Get(r, "session")
        session.Values["userID"] = userID
        session.Save(r, w)
        tpl.ExecuteTemplate(w, "index.html", "Logged In")
        return
    }
    tpl.ExecuteTemplate(w, "login.html", "Username or password is correct.")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "index.html", "Logged In")
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "about.html", "Logged In")
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	delete(session.Values, "userID")
	session.Save(r, w)
	tpl.ExecuteTemplate(w, "login.html", "Logged Out")
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
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
		tpl.ExecuteTemplate(w, "register.html", "There was a problem registering account.")
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
	userID, err := retrieveUserIDFromDB(username)
    if err != nil {
        tpl.ExecuteTemplate(w, "register.html", "There was a problem registering account.")
        return
    }
    session, _ := store.Get(r, "session")
    session.Values["userID"] = userID
    session.Save(r, w)
    http.Redirect(w, r, "/login", http.StatusFound)
}

func retrieveUserIDFromDB(username string) (string, error) {
    var userID string
    stmt := "SELECT UserID FROM bcrypt WHERE username = ?"
    row := db.QueryRow(stmt, username)
    err := row.Scan(&userID)
    if err != nil {
        if err == sql.ErrNoRows {
            return "", fmt.Errorf("user not found")
        }
        return "", err
    }
    return userID, nil
}