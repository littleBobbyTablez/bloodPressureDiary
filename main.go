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

type pageType struct {
    Page    int
    Active  bool
    OrderBy string
    Order   string
}

func main() {
    db := connectDb()

    r := gin.Default()

    r.LoadHTMLGlob("statics/*.html")

    r.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.html", gin.H{})
    })

    r.GET("/historyChart", func(c *gin.Context) {
        from := c.Query("from")
        to := c.Query("to")
        e, err := readEntries(db, 0, "t", "ASC", from, to)
        if err != nil {
            log.Fatal(err)
        }

        lab, sys, dys, puls := generateData(e)

        c.HTML(http.StatusOK, "chart.html", gin.H{"sys": sys, "dys": dys, "puls": puls, "lab": lab, "from": from, "to": to} )
    })

    r.GET("/edit", func(c *gin.Context) {

        e, err := readEntries(db, 0, "t", "ASC", "", "")
        if err != nil {
            log.Fatal(err)
        }

        var data []entryFormat
        for _, en := range e {

            data = append(data, entryFormat{en.Id, en.Sys, en.Dys, en.Puls, en.Sport, en.T.Format(time.DateOnly)})
        }

        count := getEntryCount(db, "", "")
        pages := ((count - 1) / 10) + 1

        a := make([]pageType, pages)

        for i := 1; i <= pages; i++ {
            active := i == 1
            a[i-1] = pageType{i, active, "t", "ASC"}
        }

        c.HTML(http.StatusOK, "edit.html", gin.H{"data": data, "pages": a, "active": 1, "orderBy": "t", "order": "ASC", "symbol": "9650"})
    })

    r.GET("/table/:Page", func(c *gin.Context) {
        param := c.Param("Page")
        page, err := strconv.Atoi(param)
        if err != nil {
            log.Fatal(err)
        }

        orderBy := c.Query("orderBy")
        order := c.Query("order")
        from := c.Query("from")
        to := c.Query("to")
        noswitch, parseErr := strconv.ParseBool(c.Query("noswitch"))
        if parseErr != nil {
            noswitch = false
        }
        symbol := "9650"

        if !noswitch {
            if order == "ASC" {
                order = "DESC"
            } else {
                order = "ASC"
            }
        }

        if order == "DESC" {
            symbol = "9660"
        }

        offset := (page - 1) * 10
        e, err2 := readEntries(db, offset, orderBy, order, from, to)
        if err2 != nil {
            log.Fatal(err2)
        }

        var data []entryFormat
        for _, en := range e {

            data = append(data, entryFormat{en.Id, en.Sys, en.Dys, en.Puls, en.Sport, en.T.Format(time.DateOnly)})
        }

        count := getEntryCount(db, from, to)
        pages := ((count - 1) / 10) + 1

        a := make([]pageType, pages)

        for i := 1; i <= pages; i++ {
            active := i == page
            a[i-1] = pageType{i, active, orderBy, order}
        }

        c.HTML(http.StatusOK, "edit.html", gin.H{"data": data, "pages": a, "active": page, "orderBy": orderBy, "order": order, "symbol": symbol, "from": from, "to": to})
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
    })

    r.StaticFile("/output.css", "./statics/css/output.css")

    fmt.Printf("Starting server at port 8080\n")
    if err := r.Run(":8888"); err != nil {
        log.Fatal(err)

    }
    defer db.Close()
}
