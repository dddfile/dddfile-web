package main

import (
	"dddfile/services/dataservice"
	"dddfile/util"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load(".env")
	util.CheckError(err)
	dataservice.Init()
	db := dataservice.GetDb()

	router := gin.Default()

	router.StaticFile("/android-chrome-192x192.png", "./static/android-chrome-192x192.png")
	router.StaticFile("/android-chrome-512x512.png", "./static/android-chrome-512x512.png")
	router.StaticFile("/apple-touch-icon.png", "./static/apple-touch-icon.png")
	router.StaticFile("/favicon.ico", "./static/favicon.ico")
	router.StaticFile("/favicon-16x16.png", "./static/favicon-16x16.png")
	router.StaticFile("/favicon-32x32.png", "./static/favicon-32x32.png")
	router.StaticFile("/robots.txt", "./static/robots.txt")
	router.StaticFile("/site.webmanifest", "./static/site.webmanifest")
	router.StaticFile("/sitemap.xml", "./static/sitemap.xml")

	router.Static("/assets", "./assets")

	router.SetFuncMap(template.FuncMap{
		"mod": mod,
	})
	router.LoadHTMLGlob("templates/*")

	router.GET("/", func(c *gin.Context) {
		searchText := strings.TrimSpace(c.Query("q"))
		page := c.Query("page")

		var sql string
		sqlCount := ""
		// sql := " "
		if searchText == "" {
			sql = `select id, title, url, preview_url, tags, asset_created_on from crawl_asset order by id desc LIMIT 30`
			sqlCount = `select count(*) from crawl_asset`
		} else {
			ftsPhrase := strings.Replace(searchText, "  ", " ", -1)
			for strings.Contains(ftsPhrase, "  ") {
				ftsPhrase = strings.Replace(ftsPhrase, "  ", " ", -1)
			}
			stringParts := strings.Split(ftsPhrase, " ")

			ftsPhrase = strings.Replace(ftsPhrase, " ", " <-> ", -1)
			whereClause := fmt.Sprintf("document_vectors @@ to_tsquery('%s')", ftsPhrase)
			orderClause := fmt.Sprintf(`document_vectors @@ to_tsquery('%s') desc`, ftsPhrase)
			if len(stringParts) > 0 {
				whereClause = whereClause + ` or document_vectors @@ to_tsquery('`
				orderClause = orderClause + `, document_vectors @@ to_tsquery('`
				for _, part := range stringParts {
					whereClause = whereClause + fmt.Sprintf(` %s |`, part)
					orderClause = orderClause + fmt.Sprintf(` %s |`, part)
				}
				whereClause = strings.TrimSuffix(whereClause, "|") + `')`
				orderClause = strings.TrimSuffix(orderClause, "|") + `') desc`
			}
			sql = fmt.Sprintf(`SELECT id, title, url, preview_url, tags, asset_created_on FROM crawl_asset WHERE %s ORDER BY %s LIMIT 30`, whereClause, orderClause)
			sqlCount = fmt.Sprintf(`SELECT COUNT(*) FROM crawl_asset WHERE %s`, whereClause)
		}

		if page != "" {
			pageInt, err := strconv.ParseInt(page, 10, 32)
			if pageInt > 0 {
				pageInt = pageInt - 1
			}

			if err != nil {
				panic(err)
			}
			sql = sql + " OFFSET " + strconv.FormatInt(pageInt*30, 10)
		}
		fmt.Println(sql)

		rows, err := db.Query(sql)
		util.CheckError(err)
		defer rows.Close()

		rowCount := 30
		if sqlCount != "" {
			row := db.QueryRow(sqlCount)
			err = row.Scan(&rowCount)
			util.CheckError(err)

			if rowCount > 500 {
				rowCount = 500
			}
		}

		type Asset struct {
			Id                                      int
			Title, Url, PreviewUrl, Tags, CreatedOn string
		}
		var assets []Asset
		for rows.Next() {
			var title, url, preview_url, tags, asset_created_on string
			var id int

			err = rows.Scan(&id, &title, &url, &preview_url, &tags, &asset_created_on)
			util.CheckError(err)

			// fmt.Println(title, url)

			asset := Asset{
				Id:         id,
				Title:      title,
				Url:        url,
				PreviewUrl: preview_url,
				Tags:       tags,
				CreatedOn:  asset_created_on,
			}
			assets = append(assets, asset)
		}

		util.CheckError(err)

		// var pages []int
		// pageCount := int(rowCount / 30)
		// for i := 0; i < pageCount; i++ {
		// 	pages = append(pages, i+1)
		// }

		c.HTML(http.StatusOK, "index.html", gin.H{
			"searchText": searchText,
			"assets":     assets,
			"rowCount":   rowCount,
		})
		// c.HTML(http.StatusOK, "index.tmpl", assets)
	})

	router.GET("/go/:id", func(c *gin.Context) {
		id := c.Param("id")

		sql := fmt.Sprintf(`SELECT url FROM crawl_asset WHERE id = %s`, id)
		fmt.Println(sql)

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
