package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

type Album struct {
	ID     int64   `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float32 `json:"price"`
}

var db *sql.DB

func main() {
	// Capture connection properties.
	cfg := mysql.Config{
		User:   os.Getenv("DBUSER"),
		Passwd: os.Getenv("DBPASS"),
		Net:    "tcp",
		Addr:   "127.0.0.1:3306",
		DBName: "recordings",
	}
	// Get a database handle.
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")

	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.POST("/albums", postAlbums)
	router.GET("/albums/:id", getAlbumByID)
	router.DELETE("/albums/:id", deleteAlbumByID)
	router.Run("localhost:8080")
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
	var albums []Album

	rows, _ := db.Query("SELECT * FROM album")
	defer rows.Close()
	for rows.Next() {
		var alb Album
		if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
			fmt.Println("some error")
		}
		albums = append(albums, alb)
	}
	c.IndentedJSON(http.StatusOK, albums)
}

// getAlbumByID locates the album whose ID value matches the id parameter sent by the client, then returns that album as a response.
func getAlbumByID(c *gin.Context) {
	id := c.Param("id")
	var alb Album
	row := db.QueryRow("SELECT * FROM album WHERE id = ?", id)
	if err := row.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
	} else {
		c.IndentedJSON(http.StatusOK, alb)
	}
}

// postAlbums adds an album from JSON received in the request body.
func postAlbums(c *gin.Context) {
	var alb Album
	if err := c.BindJSON(&alb); err != nil {
		fmt.Printf("addAlbum: %v", err)
	}
	result, err := db.Exec("INSERT INTO album (title, artist, price) VALUES (?, ?, ?)", alb.Title, alb.Artist, alb.Price)
	if err != nil {
		fmt.Printf("addAlbum: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		fmt.Printf("addAlbum: %v", err)
	}
	alb.ID = id
	c.IndentedJSON(http.StatusCreated, alb)
}

func deleteAlbumByID(c *gin.Context) {
	id := c.Param("id")
	var alb Album
	row := db.QueryRow("SELECT * FROM album WHERE id = ?", id)
	if err := row.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
	} else {
		db.QueryRow("DELETE FROM album WHERE id = ?", id)
		msg := fmt.Sprintf("Album %s deleted successfully", alb.Title)
		response := map[string]string{
			"message": msg,
		}
		c.IndentedJSON(http.StatusOK, response)
	}
}
