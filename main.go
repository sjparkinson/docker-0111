package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"os"
	"os/signal"
	"strings"
)

var (
	listenAddress string
)

type Doggo struct {
	Name string `json:"name"`
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	connectionString := fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s",
		os.Getenv("MYSQL_USERNAME"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_DATABASE"),
	)

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	rows, err := db.Query("SELECT name FROM doggos")

	if err != nil {
		panic(err)
	}
	
	defer rows.Close()

	type dog struct {
		name string
	}

	dogs := []dog{}

	for rows.Next() {
		var d dog

		if err = rows.Scan(&d.name); err != nil {
			panic(err)
		}

		dogs = append(dogs, d)
	}

	var list strings.Builder

	for _, d := range dogs {
		fmt.Fprintf(&list, "<li>%s</li>\n", d.name)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fmt.Sprintf(`<!DOCTYPE html>
	<html lang="en-GB">
		<head>
		<meta charset="UTF-8">
		<meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1">
		<meta name="viewport" content="width=device-width,initial-scale=1">
		<title>Docker 0111 – Doggos</title>
		</head>
		<body>
		<h1>Docker 0111 – Doggos</h1>
		<p>
			<ul>
				%s
			</ul>
		</p>
		</body>
	</html>`, list.String())))
}

func main() {
	flag.StringVar(&listenAddress, "http", ":8080", "Address to listen on for the web interface.")
	flag.Parse()

	router := http.NewServeMux()

	router.Handle("/", http.HandlerFunc(index))

	server := &http.Server{
		Addr:    listenAddress,
		Handler: router,
	}

	done := make(chan bool)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, os.Interrupt)

		<-quit

		if err := server.Close(); err != nil {
			panic(err)
		}

		close(done)
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}

	fmt.Printf("Started listening at %s", listenAddress)

	<-done
}
