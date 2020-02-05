package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gorilla/mux"
)

type Body struct {
	Message  string   `json:"message"`
	Employee []Person `json:"employee"`
}

type Person struct {
	ID     uint16 `json:"id"`
	Name   string `json:"name"`
	Salary uint32 `json:"salary"`
}

type Conf map[string]map[string]interface{}

const jsonData string = `
{
	"message": "%s",
	"employee": [%s]
}
`

var (
	endPoint   string = LoadConf("db")
	myServer   string = GetSelfConf("was")
	client            = &http.Client{}
	dbEndpoint        = GetDbEndpoint()
)

func GetSelfConf(target string) string {
	var hostString string
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			hostString = addr.String()
			break
		}
	}

	f, err := os.Open("../conf.json")
	if err != nil {
		log.Fatal(err)
	}

	raw, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	var conf Conf
	json.Unmarshal(raw, &conf)

	return fmt.Sprintf("%s:%d", hostString, int(conf["port"][target].(float64)))
}

func LoadConf(target string) string {

	f, err := os.Open("../conf.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	raw, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	var conf Conf
	json.Unmarshal(raw, &conf)
	return conf["host"][target].(string) + ":" + strconv.Itoa(int(conf["port"][target].(float64)))
}

func GetDbEndpoint() string {
	userName := os.Getenv("DBUSERNAME") 
	userPass := os.Getenv("DBPASSWORD") 
	dbName := os.Getenv("DBNAME")       

	dbEndpoint := fmt.Sprintf("%s:%s@tcp(%s)/%s", userName, userPass, endPoint, dbName)

	return dbEndpoint
}

func ParseURL(url string) []string {
	start, index := 0, 0
	parse := []rune(url)[1:]
	result := []string{}

	for index <= len(parse)-1 {

		if parse[index] == rune('/') {
			result = append(result, string(parse[start:index]))
			if index != len(parse)-1 {
				start = index + 1
			}
		} else if index == len(parse)-1 {
			result = append(result, string(parse[start:]))
		}

		index++
	}
	return result
}

func Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	str := `
		Human Resources v1.0
	`
	message := fmt.Sprintf(jsonData, str, "")
	w.Write([]byte(message))
}

func GetAll(w http.ResponseWriter, r *http.Request) {
	fmt.Println("REQUEST RECHEAD!")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	db, err := sql.Open("mysql", dbEndpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var result []Person
	var person Person

	rows, err := db.Query("SELECT id, name, salary FROM hr")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&(person.ID), &(person.Name), &(person.Salary))
		if err != nil {
			log.Fatal(err)
		}
		result = append(result, person)
	}

	var stringJsonify string
	
	if len(result) != 0 {
		byteJsonify, err := json.Marshal(result)
		if err != nil {
			log.Fatal(err)
		}
		stringJsonify = string(byteJsonify)[1 : len(byteJsonify)-1]
	} else {
		stringJsonify = ""
	}

	message := fmt.Sprintf(jsonData, "Successfully retreived HR resources", stringJsonify)
	w.Write([]byte(message))
}

func GetSpecific(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	fields := ParseURL(r.URL.Path)
	var field string = fields[1]
	var value string = fields[2]

	dbEndpoint := GetDbEndpoint()

	db, err := sql.Open("mysql", dbEndpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var result []Person
	var person Person

	if field == "name" {
		value = "'" + value + "'"
	}

	queryString := fmt.Sprintf("SELECT id, name, salary FROM hr WHERE %s = %s", field, value)
	
	rows, err := db.Query(queryString)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&(person.ID), &(person.Name), &(person.Salary))
		if err != nil {
			log.Fatal(err)
		}
		result = append(result, person)
	}

	var stringJsonify string

	if len(result) != 0 {
		byteJsonify, err := json.Marshal(result)
		if err != nil {
			log.Fatal(err)
		}
		stringJsonify = string(byteJsonify)[1 : len(byteJsonify)-1]
	} else {
		stringJsonify = ""
	}

	message := fmt.Sprintf(jsonData, "Successfully retreived HR resources", stringJsonify)
	w.Write([]byte(message))
}

func Post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rawJson := make([]byte, r.ContentLength)
	r.Body.Read(rawJson)

	var response Body
	json.Unmarshal(rawJson, &response)

	if len(response.Employee) < 1 {
		message := fmt.Sprintf(jsonData, "Invalid input", "")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(message))
	}

	db, err := sql.Open("mysql", dbEndpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	nRows := 0
	fmt.Println("oioi")
	for _, person := range response.Employee {
		execString := fmt.Sprintf("INSERT INTO hr (id, name, salary) value (%d, '%s', %d)", person.ID, person.Name, person.Salary)
		result, err := db.Exec(execString)
		if err != nil {
			log.Fatal(err)
		}
		n, err := result.RowsAffected()
		nRows += int(n)
	}
	fmt.Println("ioio")
	w.WriteHeader(http.StatusCreated)
	insideMessage := fmt.Sprintf("%d ROWS CREATED!", nRows)
	message := fmt.Sprintf(jsonData, insideMessage, "")
	w.Write([]byte(message))
}

func Put(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	fields := ParseURL(r.URL.Path)

	field := fields[1]
	value := fields[2]

	rawJson := make([]byte, r.ContentLength)
	var parsedJson Body

	r.Body.Read(rawJson)
	err := json.Unmarshal(rawJson, &parsedJson)
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("mysql", dbEndpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if field == "name" {
		value = "'" + value + "'"
	}
	execString := fmt.Sprintf("UPDATE hr SET id = %d, name = '%s', salary = %d where %s = %s", parsedJson.Employee[0].ID, parsedJson.Employee[0].Name, parsedJson.Employee[0].Salary, field, value)

	result, err := db.Exec(execString)
	if err != nil {
		log.Fatal(err)
	}

	nRows, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	insideMessage := fmt.Sprintf("%d ROWS AFFECTED!", nRows)
	message := fmt.Sprintf(jsonData, insideMessage, "")
	w.Write([]byte(message))

}

func Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	fields := ParseURL(r.URL.Path)

	field := fields[1]
	value := fields[2]

	db, err := sql.Open("mysql", dbEndpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if field == "name" {
		value = "'" + value + "'"
	}

	result, err := db.Exec("DELETE FROM hr WHERE ? = ?", field, value)
	if err != nil {
		log.Fatal(err)
	}

	nRows, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}

	insideMessage := fmt.Sprintf("%d ROWS AFFECTED!", nRows)
	message := fmt.Sprintf(jsonData, insideMessage, "")
	w.Write([]byte(message))
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	message := fmt.Sprintf(jsonData, "REQUEST INVALID!", "")
	w.Write([]byte(message))
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", Index).Methods(http.MethodGet)

	HR := r.PathPrefix("/employee").Subrouter()

	HR.HandleFunc("", GetAll).Methods(http.MethodGet)
	// HR.HandleFunc("", NotFound)

	HR.HandleFunc("/new", Post).Methods(http.MethodPost)
	HR.HandleFunc("/new", NotFound)

	HR.HandleFunc("/id/{targetID}", GetSpecific).Methods(http.MethodGet)
	HR.HandleFunc("/id/{targetID}", Put).Methods(http.MethodPut)
	HR.HandleFunc("/id/{targetID}", Delete).Methods(http.MethodDelete)
	HR.HandleFunc("/id", NotFound)

	HR.HandleFunc("/name/{targetName}", GetSpecific).Methods(http.MethodGet)
	HR.HandleFunc("/name/{targetName}", Put).Methods(http.MethodPut)
	HR.HandleFunc("/name/{targetName}", Delete).Methods(http.MethodDelete)
	HR.HandleFunc("/name", NotFound)

	log.Fatal(http.ListenAndServe(myServer, r))
}