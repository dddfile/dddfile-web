package main

import (
  "net/http"
  "github.com/gin-gonic/gin"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	"os"
)

func main() {
	err := godotenv.Load(".env")
	CheckError(err)

	var (
		host     = getEnvVar("DATABASE_HOST")
		port     = getEnvVar("DATABASE_PORT")
		username = getEnvVar("DATABASE_USERNAME")
		password = getEnvVar("DATABASE_PASSWORD")
		database = getEnvVar("DATABASE_NAME")
	)

  router := gin.Default()
	router.Static("/assets", "./assets")
	router.LoadHTMLGlob("templates/*")
  router.GET("/", func(c *gin.Context) {

		
		// connection string
    psqlconn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", host, port, username, password, database)
		if getEnvVar("GIN_MODE") != "release" {
			psqlconn = psqlconn + " sslmode=disable"
		}
		
		// open database
    db, err := sql.Open("postgres", psqlconn)
    CheckError(err)
		
    // close database
    defer db.Close()
		
		searchText := c.Query("q")
		var sql string
		// sql := " "
		if searchText == "" {
			sql = `select title, url, preview_url, tags, asset_created_on from crawl_asset`
		} else {
			sql = fmt.Sprintf(`SELECT title, url, preview_url, tags, asset_created_on FROM crawl_asset WHERE document_vectors @@ to_tsquery('%s')`, searchText)
		}
			
		fmt.Println(sql);
		rows, err := db.Query(sql)
		CheckError(err)
		defer rows.Close()
		
		type Asset struct {
			Title, Url, PreviewUrl, Tags, CreatedOn string
		}
		var assets []Asset
		for rows.Next() {
				var title, url, preview_url, tags, asset_created_on string
		
				err = rows.Scan(&title, &url, &preview_url, &tags, &asset_created_on)
				CheckError(err)
		
				// fmt.Println(title, url)
				
				asset := Asset{
					Title: title,
					Url: url,
					PreviewUrl: preview_url,
					Tags: tags,
					CreatedOn: asset_created_on,
				}
				assets = append(assets, asset)
		}
		
		CheckError(err)

		// c.HTML(http.StatusOK, "index.tmpl", gin.H {
    //   "title": "Main website",
    // })
		c.HTML(http.StatusOK, "index.html", gin.H {
			"searchText": searchText,
			"assets": assets,
		})
		// c.HTML(http.StatusOK, "index.tmpl", assets)
  })
  router.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func CheckError(err error) {
	if err != nil {
			panic(err)
	}
}

func getEnvVar(key string) string {
	return os.Getenv(key)
}