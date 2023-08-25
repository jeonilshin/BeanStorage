package main

import (
	"database/sql"
	"encoding/gob"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID	   	  string
	Username  string
	Email     string
	pswdHash  string
	CreatedAt string
	Active    string
	verHash   string
	timeout   string
}

var db *sql.DB

var store = sessions.NewCookieStore([]byte("super-secret"))

func init() {
	store.Options.HttpOnly = true
	store.Options.Secure = true
	gob.Register(&User{})
}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	var err error
	db, err = sql.Open("mysql", "shin:Skills2024**@tcp(beanstorage.cmu7dpsydrgb.us-east-1.rds.amazonaws.com:3306)/bean_storage")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	autoRouter := router.Group("/user", auth)

	router.GET("/", indexHandler)
	router.GET("/login", loginGETHandler)
	router.POST("/login", loginPOSTHandler)

	authRouter.GET("/profile", profileHandler)

	err = router.Run("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
}

func auth(c *gin.Context) {
	session, _ := store.Get(c.Request, "session")
	_, ok := session.Values["user"]
	if !ok {
		c.HTML(http.StatusForbidden, "login.html", nil)
		c.Abort()
	}
	c.Next()
}

func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func loginGETHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func loginPOSThandler(c *gin.Context) {
	var user User
	user.Username = c.PostForm("username")
	password := c.PostForm("password")
	err := user.getUserByUsername()
	if err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"message": "check usernane or password"})
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.pswdHash), []byte(password))
	if err != nil {
		session, _ := store.Get(c.Request, "session")
		session.Values["user"] = user
		session.Save(c.Request, c.Writer)
		c.HTML(http.StatusOK, "loggedin.html", gin.H{"username": user.Username})
		return
	}
	c.HTML(http.StatusUnauthorized, "login.html", gin.H{"message": "check username or password"})
}

func profileHandler(c *gin.Context) {
	session, _ := store.Get(c.Request, "session")
	var user = &User{}
	val := session.Values["user"]
	var ok bool
	if user, ok = val.(*User); !ok {
		c.HTML(http.StatusUnauthorized, "login.html", nil)
		return
	}
	c.HTML(http.StatusOK, "profile.html", gin.H{"user": user})
}

func (u *User) getUserByUsername() error {
	stmt := "SELECT * FROM users WHERE username = ?"
	row := db.QueryRow(stmt, u.Username)
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.pswdHash, &u.CreatedAt, &u.Active, &u.verHash, &u.timeout)
	if err != nil {
		return err
	}
	return nil
}