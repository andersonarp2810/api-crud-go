package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite3", "test.db")
	if err != nil {
		log.Fatal("Falha ao conectar db")
	}
	st, _ := db.Prepare("CREATE TABLE IF NOT EXISTS people (id INTEGER PRIMARY KEY AUTOINCREMENT, firstname TEXT, lastname TEXT)")
	st.Exec()
}

func main() {

	router := gin.Default()
	v1 := router.Group("/api/v1/people")
	{
		v1.POST("/", createPerson)
		v1.GET("/", fetchAllPeople)
		v1.PUT("/:id", updatePerson)
		v1.GET("/:id", getPerson)
		v1.DELETE("/:id", deletePerson)
	}
	router.Run()
}

type (
	personModel struct {
		FirstName string `json:"firstname"`
		LastName  string `json:"lastname"`
	}
	transformedPerson struct {
		ID        uint   `json:"id"`
		FirstName string `json:"firstname"`
		LastName  string `json:"lastname"`
	}
)

func createPerson(c *gin.Context) {
	/*
		st, _ = db.Prepare("INSERT INTO people (firstname, lastname) VALUES (?, ?)")
		st.Exec("po", "teito")
	*/
	person := personModel{FirstName: c.PostForm("firstname"), LastName: c.PostForm("lastname")}
	log.Println(person, person.FirstName, person.LastName)
	sql, args, _ := sq.Insert("people").Columns("firstname", "lastname").Values(person.FirstName, person.LastName).ToSql()
	log.Println(sql)
	log.Println(args)
	st, err := db.Prepare(sql)
	if err != nil {
		log.Fatal(err)
	}
	st.Exec(args...)

	c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated})
}

func fetchAllPeople(c *gin.Context) {
	//var people []personModel
	var _people []transformedPerson
	var id uint
	var firstname string
	var lastname string

	sql, _, _ := sq.Select("id", "firstname", "lastname").From("people").ToSql()
	log.Println(sql)
	rows, err := db.Query(sql)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&id, &firstname, &lastname); err != nil {
			log.Fatal(err)
		}
		person := transformedPerson{ID: id, FirstName: firstname, LastName: lastname}
		_people = append(_people, person)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": _people})

}

func updatePerson(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var firstname string = c.PostForm("firstname")
	var lastname string = c.PostForm("lastname")
	log.Println(id, firstname, lastname)

	pre := sq.Update("people")
	if firstname != "" {
		pre = pre.Set("firstname", firstname)
	}
	if lastname != "" {
		pre = pre.Set("lastname", lastname)
	}
	sql, args, err := pre.Where("id = ?", id).ToSql()
	if err != nil {
		log.Fatal(err)
		return
	}

	st, err := db.Prepare(sql)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println(sql, args)
	st.Exec(args...)

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK})

}

func getPerson(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	sql, args, err := sq.Select("firstname", "lastname").From("people").Where("id = ?", id).ToSql()

	rows, err := db.Query(sql, args...)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer rows.Close()

	var firstname, lastname string
	var person transformedPerson

	for rows.Next() {
		if err := rows.Scan(&firstname, &lastname); err != nil {
			log.Fatal(err)
			return
		}
		person = transformedPerson{ID: uint(id), FirstName: firstname, LastName: lastname}
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
		return
	}

	if person.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "data": person})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": person})

}

func deletePerson(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	sql, args, err := sq.Delete("people").Where("id = ?", id).ToSql()

	st, err := db.Prepare(sql)

	if err != nil {
		log.Fatal(err)
	}

	res, err := st.Exec(args...)

	if err != nil {
		log.Fatal(err)
	}

	ra, err := res.RowsAffected()

	if err != nil {
		log.Fatal(err)
	}

	if ra == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": fmt.Sprint("ID não encontrado.")})
	} else if ra != 1 {
		log.Fatal("Mais de uma linha afetada")
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK})

}
