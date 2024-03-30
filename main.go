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
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

func main() {
	cfg := mysql.Config{
		User:	os.Getenv("DBUSER"),
		Passwd:	os.Getenv("DBPASS"),
		Net:	"tcp",
		Addr:	"127.0.0.1:3306",
		DBName:	"recordings",
	}

	var db *sql.DB
	var err error

	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal("ERR : ", err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal("PING : ", pingErr)
	}

	fmt.Println("Connected !")

	router := gin.Default()
	router.GET("/albums", func(c *gin.Context) {
		getAlbums(c, db)
	})
	router.GET("/albums/:id", func(c * gin.Context) {
		getAlbumByID(c, db)
	})
	router.POST("/albums", func(c *gin.Context) {
		postAlbums(c, db, Album{
			Title: "The Modern Sound of Betty Carter",
			Artist: "Betty Carter",
			Price: 49.99,
		})
	})

	router.Run("localhost:8000")
}

func getAlbums(c *gin.Context, db *sql.DB) {
	rows, err := db.Query("SELECT * FROM album")
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var albums []Album

	for rows.Next() {
		var album Album
		if err := rows.Scan(&album.ID, &album.Title, &album.Artist, &album.Price); err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		albums = append(albums, album)
	}

	if err := rows.Err(); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.IndentedJSON(http.StatusOK, albums)
}

func getAlbumByID(c *gin.Context, db *sql.DB) {
	id := c.Param("id")
	var album Album

	row := db.QueryRow("SELECT * FROM album WHERE id = ?", id)
	if err := row.Scan(&album.ID, &album.Title, &album.Artist, &album.Price); err != nil {
		if err == sql.ErrNoRows {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.IndentedJSON(http.StatusOK, album)
}

func postAlbums(c *gin.Context, db *sql.DB, newAlbum Album) {
	result, err := db.Exec("INSERT INTO album (title, artist, price) VALUES (?, ?, ?)", newAlbum.Title, newAlbum.Artist, newAlbum.Price)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.IndentedJSON(http.StatusCreated, id)
}
