package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

type entry struct {
	Id    int
	Sys   int64
	Dys   int64
	Puls  int64
	Sport bool
	T     time.Time
}

type entryFormat struct {
	Id    int
	Sys   int64
	Dys   int64
	Puls  int64
	Sport bool
	T     string
}

type values struct {
	Sys  int64
	Dys  int64
	Puls int64
	T    string
}

func main() {
	db := connectDb()

	r := gin.Default()

	r.LoadHTMLGlob("statics/*.html")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.GET("/historyChart", func(c *gin.Context) {
		e, err := readAllEntries(db)
		if err != nil {
			log.Fatal(err)
		}

		lab, sys, dys, puls := generateData(e)

		c.HTML(http.StatusOK, "chart.html", gin.H{"sys": sys, "dys": dys, "puls": puls, "lab": lab})
	})

	r.GET("/edit", func(c *gin.Context) {

		count := getEntryCount(db)

		e, err := readEntries(db, 0)
		if err != nil {
			log.Fatal(err)
		}

		var data []entryFormat
		for _, en := range e {

			data = append(data, entryFormat{en.Id, en.Sys, en.Dys, en.Puls, en.Sport, en.T.Format(time.DateOnly)})
		}

		pages := ((count - 1) / 10) + 1

		a := make([]int, pages)

		for i := 1; i <= pages; i++ {
			a[i-1] = i
		}

		c.HTML(http.StatusOK, "edit.html", gin.H{"data": data, "pages": a})
	})

    r.GET("/table/:Page", func(c *gin.Context) {
        param := c.Param("Page")
        page, err := strconv.Atoi(param)
        if err != nil {
            log.Fatal(err)
        }
        
        offset := (page - 1) * 10
        e, _ := readEntries(db, offset)
        
        var data []entryFormat
		for _, en := range e {

			data = append(data, entryFormat{en.Id, en.Sys, en.Dys, en.Puls, en.Sport, en.T.Format(time.DateOnly)})
		}
 
        c.HTML(http.StatusOK, "table.html", gin.H{"data": data})
    })

	r.GET("/entry/:Id", func(c *gin.Context) {
		param := c.Param("Id")
		id, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		var e entry
		e, err = readEntry(db, int(id))

		c.HTML(http.StatusOK, "entry.html", gin.H{"entry": entryFormat{e.Id, e.Sys, e.Dys, e.Puls, e.Sport, e.T.Format(time.DateOnly)}})
	})

	r.PUT("/entry/:Id", func(c *gin.Context) {
		param := c.Param("Id")
		id, err := strconv.ParseInt(param, 10, 64)
		s, _ := strconv.ParseInt(c.Request.FormValue("sys"), 10, 64)
		d, _ := strconv.ParseInt(c.Request.FormValue("dys"), 10, 64)
		p, _ := strconv.ParseInt(c.Request.FormValue("puls"), 10, 64)
		sp, _ := strconv.ParseBool(c.Request.FormValue("sport"))

		if err != nil {
			log.Fatal(err)
		}

		var e entry
		e, err = readEntry(db, int(id))

		e.Sys = s
		e.Dys = d
		e.Sport = sp
		e.Puls = p

		updateEntry(e, db)

		c.HTML(http.StatusOK, "entry.html", gin.H{"entry": entryFormat{e.Id, e.Sys, e.Dys, e.Puls, e.Sport, e.T.Format(time.DateOnly)}})
	})

	r.GET("/editrow/:Id", func(c *gin.Context) {
		param := c.Param("Id")
		id, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		var e entry
		e, err = readEntry(db, int(id))

		output := entryFormat{e.Id, e.Sys, e.Dys, e.Puls, e.Sport, e.T.Format(time.DateOnly)}

		c.HTML(http.StatusOK, "editrow.html", gin.H{"entry": output})
	})

	r.GET("/addnew", func(c *gin.Context) {
		c.HTML(http.StatusOK, "addnew.html", gin.H{})
	})

	r.POST("/add", func(c *gin.Context) {
		s, _ := strconv.ParseInt(c.PostForm("sys"), 10, 64)
		d, _ := strconv.ParseInt(c.PostForm("dys"), 10, 64)
		p, _ := strconv.ParseInt(c.PostForm("puls"), 10, 64)
		sp, _ := strconv.ParseBool(c.PostForm("sport"))

		e := entry{
			0,
			s,
			d,
			p,
			sp,
			time.Now(),
		}

		writeEntry(e, db)

		c.HTML(http.StatusOK, "success.html", gin.H{})
	})

	r.StaticFile("/output.css", "./statics/css/output.css")

	fmt.Printf("Starting server at port 8080\n")
	if err := r.Run(":8888"); err != nil {
		log.Fatal(err)

	}
	defer db.Close()
}
