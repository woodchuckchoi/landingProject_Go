package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"os"
	"io/ioutil"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"encoding/json"
	"net"

	"github.com/gorilla/mux"
)

type Body	struct {
	Message		string		`json:"message"`
	Employee	[]Person	`json:"employee"`
}

type Person struct {
	ID		uint16	`json:"id"`
	Name	string	`json:"name"`
	Salary	uint32	`json:"salary"`
}

type Conf map[string] map[string]interface{}

const jsonData string = `
{
	"message": "%s",
	"employee": [%s]
}
`

var (
	endPoint string = LoadConf("db")
	myServer string	= "http://" + GetSelfConf("was")
	client 			= &http.Client{}
	dbEndpoint 		= GetDbEndpoint()
)

func GetSelfConf(target string) string{
	f, err := os.Open("../conf.json")
	if err != nil {
		log.Fatal(err)
	}

	raw, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	var to_return string
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			to_return = fmt.Sprintf("%v", ipv4)
		}   
	}

	var conf Conf
	json.Unmarshal(raw, &conf)
	return to_return + ":" + strconv.Itoa(int(conf["port"][target].(float64)))
}

func LoadConf(target string) string {
	
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
	return conf["host"][target].(string) + ":" + strconv.Itoa(int(conf["port"][target].(float64)))
}

func GetDbEndpoint() string {
	userName	:= os.Getenv("DBUSERNAME") // master
	userPass	:= os.Getenv("DBPASSWORD") // Bespin1!
	dbName		:= os.Getenv("DBNAME") // hr

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
			if index != len(parse) - 1 {
				start = index + 1
			}
		} else if index == len(parse) - 1 {
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
	
	byteJsonify, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}
	stringJsonify := string(byteJsonify)[1:len(byteJsonify)-1]
	message := fmt.Sprintf(jsonData, "Successfully retreived HR resources", stringJsonify)
	w.Write([]byte(message))

	// req, err := http.NewRequest("GET", endPoint, nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// resp, err := client.Do(req)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer resp.Close()
	
	// bytes, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// w.Write(bytes)
}

func GetSpecific(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	fields := ParseURL(r.URL.Path)
	var field string = fields[1]
	var value interface{} = fields[2]

	dbEndpoint := GetDbEndpoint()

	db, err := sql.Open("mysql", dbEndpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var result []Person
	var person Person
	
	rows, err := db.Query("SELECT id, name, salary FROM hr WHERE ? = ?", field, value)
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
	
	byteJsonify, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}
	stringJsonify := string(byteJsonify)[1:len(byteJsonify)-1]
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

	for _, person := range(response.Employee) {
		result, err := db.Exec("INSERT INTO hr (id, name, salary) value (?, '?', ?)", person.ID, person.Name, person.Salary)
		if err != nil {
			log.Fatal(err)
		}
		n, err := result.RowsAffected()
		nRows += int(n)
	}

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

	result, err := db.Exec("UPDATE hr SET id = ?, name = '?', salary = ? where ? = ?", parsedJson.Employee[0].ID, parsedJson.Employee[0].Name, parsedJson.Employee[0].Salary, field, value)
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
	HR.HandleFunc("", NotFound)

	HR.HandleFunc("/new", Post).Methods(http.MethodPost)
	HR.HandleFunc("/new", NotFound)

	HR.HandleFunc("/id", GetSpecific).Methods(http.MethodGet)
	HR.HandleFunc("/id", Put).Methods(http.MethodPut)
	HR.HandleFunc("/id", Delete).Methods(http.MethodDelete)
	HR.HandleFunc("/id", NotFound)

	HR.HandleFunc("/name", GetSpecific).Methods(http.MethodGet)
	HR.HandleFunc("/name", Put).Methods(http.MethodPut)
	HR.HandleFunc("/name", Delete).Methods(http.MethodDelete)
	HR.HandleFunc("/name", NotFound)

	log.Fatal(http.ListenAndServe(myServer, r))
}


// layer1 = display received data / request json				inbound: N/A | outbound: 777
// layer2 = forming to-be-displayed data / relay request		inbound: 777 | outbound: 888
// layer3 = parse request / communicate with db				inbound: 888 | outbound: DB(3906)