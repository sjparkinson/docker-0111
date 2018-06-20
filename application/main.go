package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"strings"
)

var (
	listenAddress string
	db *sql.DB
)

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	rows, err := db.Query("SELECT name FROM doggos")

	if err != nil {
		log.Error(err)

		w.WriteHeader(http.StatusServiceUnavailable)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<!DOCTYPE html>
		<html lang="en-GB">
			<head>
			<meta charset="UTF-8">
			<meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1">
			<meta name="viewport" content="width=device-width,initial-scale=1">
			<title>Docker 0111 – Doggos 🐶</title>
			</head>
			<body>
			<h1>Docker 0111 – Doggos 🐶</h1>
			<p>Sorry, MySQL is not ready yet!</p>
			</body>
		</html>`))

		return
	}

	defer rows.Close()

	type dog struct {
		name string
	}

	dogs := []dog{}

	for rows.Next() {
		var d dog

		if err = rows.Scan(&d.name); err != nil {
			log.Fatal(err)
		}

		dogs = append(dogs, d)
	}

	log.Infof("found %d dogs in MySQL", len(dogs))

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
		<title>Docker 0111 – Doggos 🐶</title>
		</head>
		<body>
		<h1>Docker 0111 – Doggos 🐶</h1>
		<p>Here's our list of dogs, we found %d in total!</p>
		<ul>
			%s
		</ul>
		</body>
	</html>`, len(dogs), list.String())))
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
			log.Fatal(err)
		}

		close(done)
	}()

	connectionString := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s",
		os.Getenv("MYSQL_USERNAME"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_ADDRESS"),
		os.Getenv("MYSQL_DATABASE"),
	)

	var err error
	db, err = sql.Open("mysql", connectionString)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	log.Infof("connected to MySQL at %s", os.Getenv("MYSQL_ADDRESS"))

	log.WithFields(log.Fields{
		"address": listenAddress,
	}).Info("started listening")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-done

	log.Info("stopped")
}
