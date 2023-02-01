package main

import (
  "net/http"
  "github.com/gin-gonic/gin"
	"database/sql"
	"fmt"
	"html/template"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	"os"
	"strings"
	"strconv"
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

  router := gin.Default()
	
	router.StaticFile("/sitemap.xml", "./static/sitemap.xml")
	router.StaticFile("/robots.txt", "./static/robots.txt")
	router.Static("/assets", "./assets")

	router.SetFuncMap(template.FuncMap{
		"mod": mod,
	})
	router.LoadHTMLGlob("templates/*")
  router.GET("/", func(c *gin.Context) {		
		searchText := c.Query("q")
		page := c.Query("page")

		var sql string
		sqlCount := ""
		// sql := " "
		if searchText == "" {
			sql = `select id, title, url, preview_url, tags, asset_created_on from crawl_asset order by id desc LIMIT 30`
			sqlCount = `select count(*) from crawl_asset`
		} else {
			ftsPhrase := "'" + strings.Replace(searchText, " ", " <-> ", 10) + "'"
			sql = fmt.Sprintf(`SELECT id, title, url, preview_url, tags, asset_created_on FROM crawl_asset WHERE document_vectors @@ to_tsquery(%s) order by id desc  LIMIT 30`, ftsPhrase)
			sqlCount = fmt.Sprintf(`SELECT COUNT(*) FROM crawl_asset WHERE document_vectors @@ to_tsquery(%s)`, ftsPhrase)
		}

		if page != "" {
			pageInt, err := strconv.ParseInt(page, 10, 32)
			if (pageInt > 0) {
				pageInt = pageInt - 1
			}

			if err != nil {
				panic(err)
			}
			sql = sql + " OFFSET " + strconv.FormatInt(pageInt * 30, 10)
		}
			
		fmt.Println(sql);
		rows, err := db.Query(sql)
		CheckError(err)
		defer rows.Close()

		rowCount := 30
		if sqlCount != "" {
			row := db.QueryRow(sqlCount)
			err = row.Scan(&rowCount)
			CheckError(err)

			if rowCount > 500 {
				rowCount = 500
			}
		}
		
		type Asset struct {
			Id int
			Title, Url, PreviewUrl, Tags, CreatedOn string
		}
		var assets []Asset
		for rows.Next() {
				var title, url, preview_url, tags, asset_created_on string
				var id int
		
				err = rows.Scan(&id, &title, &url, &preview_url, &tags, &asset_created_on)
				CheckError(err)
		
				// fmt.Println(title, url)
				
				asset := Asset{
					Id: id,
					Title: title,
					Url: url,
					PreviewUrl: preview_url,
					Tags: tags,
					CreatedOn: asset_created_on,
				}
				assets = append(assets, asset)
		}
		
		CheckError(err)

		// var pages []int
		// pageCount := int(rowCount / 30)
		// for i := 0; i < pageCount; i++ {
		// 	pages = append(pages, i+1)
		// }

		c.HTML(http.StatusOK, "index.html", gin.H {
			"searchText": searchText,
			"assets": assets,
			"rowCount": rowCount,
		})
		// c.HTML(http.StatusOK, "index.tmpl", assets)
  })

	router.GET("/go/:id", func(c *gin.Context) {
		id := c.Param("id")

		sql := fmt.Sprintf(`SELECT url FROM crawl_asset WHERE id = %s`, id)
		fmt.Println(sql);
		
		var url string
		row := db.QueryRow(sql)
		err = row.Scan(&url)
		if err != nil {
			c.Redirect(http.StatusFound, "/")
		} else {
			c.Redirect(http.StatusFound, url)
		}
	})

  router.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func mod(i, j int) bool { return i%j == 0 }

func CheckError(err error) {
	if err != nil {
			panic(err)
	}
}

func getEnvVar(key string) string {
	return os.Getenv(key)
}