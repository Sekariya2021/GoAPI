package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var (
	db *sql.DB
)

type Config struct {
	Dbaddress  string `json:"dbaddress"`
	Dpassword  string `json:"dbpassword"`
	Dbname     string `json:"dbname"`
	Dbport     int16  `json:"dbport"`
	Dbtable    string `json:"dbtable"`
	Dbusername string `json:"dbusername"`
	Httpport   int16  `json:"httpport"`
}

func main() {
	fmt.Println("DHT11 Sensor API")

	// Leest de config file.
	fmt.Println("Config-file aan het lezen..")
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("Bestand geopend.")
	defer file.Close()
	result, _ := ioutil.ReadAll(file)
	// Variabelen ophalen uit config
	var config Config
	json.Unmarshal(result, &config)
	username := config.Dbusername
	password := config.Dpassword
	address := config.Dbaddress
	dbport := config.Dbport
	dbName := config.Dbname
	port := config.Httpport

	// connect to database
	fmt.Printf("Verbinding maken met MySQL-server: %s:%d\n", address, dbport)
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", username, password, address, dbport, dbName))
	if err != nil {
		fmt.Printf("Connectie met de mysql server error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("MySQL-server succesvol verbonden.")

	fmt.Println("HTTP API-Server gestart op poort :" + fmt.Sprintf(" %d", port))
	// setup http server
	http.HandleFunc("/create", config.handleDataReceivedRequest)
	// start listening http request
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Printf("Fout bij luisteren naar http-verzoek: %v\n", err)
		os.Exit(1)
	}
}

// Data struct from ESP32
type Data struct {
	Temperature float32 `json:"temperature"`
	Humidity    float32 `json:"humidity"`
}

// handleDataReceivedRequest handle http requests
func (config *Config) handleDataReceivedRequest(w http.ResponseWriter, r *http.Request) {
	// return als request niet post is
	if r.Method != "POST" {
		_, err := fmt.Fprintln(w, "API kan alleen worden aangeroepen door POST Request")
		if err != nil {
			fmt.Printf("Fout bij schrijven naar http-client: %v\n", err)
		}
		return
	}
	// leez request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Fout bij het lezen van HTTP-verzoektekst: %v\n", err)
		return
	}
	// Grijpt data van request body
	data := Data{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Printf("Unmarshal json van http-verzoek body error: %v\n", err)
		return
	}
	// insert data into db
	stmt, err := db.Prepare("INSERT INTO " + config.Dbtable + " (temperature,humidity) VALUES(?,?);")
	if err != nil {
		fmt.Printf("Fout bij het voorbereiden van mysql-query: %v\n", err)
		return
	}

	// execute query
	_, err = stmt.Exec(data.Temperature, data.Humidity)
	if err != nil {
		fmt.Printf("Mysql-queryfout uitvoeren: %v\n", err)
		return
	}
	fmt.Fprintln(w, "Verbonden.")
}
