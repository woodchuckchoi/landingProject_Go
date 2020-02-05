package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

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
	endPoint string = "http://" + LoadConf("was")
	myServer string = GetSelfConf("web")
	client          = &http.Client{}
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

func ParseBody(raw []byte) []byte {
	var body Body

	pre_parse := []byte(strings.TrimSpace(string(raw)))

	err := json.Unmarshal(pre_parse, &body)
	if err != nil {
		log.Fatal(err)
	}
	
	rawResultFormat := `
	STATUS:%s
	%s
	`

	PersonFormat := `
	Entry %d |   ID %05d   | NAME %20s | SALARY %10d
	`
	
	personSum := ""
	for index, person := range body.Employee {
		personSum += fmt.Sprintf(PersonFormat, index, person.ID, person.Name, person.Salary)
	}
	
	rawResult := fmt.Sprintf(rawResultFormat, body.Message, personSum)
	
	return []byte(rawResult)
}

func Receive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	var body Body
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()

	err = json.Unmarshal(raw, &body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("RECEIVED REQUEST FROM ", r.Host, "\n", body)

	req, err := http.NewRequest(r.Method, body.Message, bytes.NewBuffer(raw))
	if err != nil {
		log.Fatal(err)
	}
	
	resp, err := client.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("BAD REQUEST!"))
		return
	}
	defer resp.Body.Close()
	
	w.WriteHeader(http.StatusOK)
	bytes, err := ioutil.ReadAll(resp.Body)
	
	w.Write(ParseBody(bytes))
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", Receive)
	log.Fatal(http.ListenAndServe(myServer, r))
}

// layer1 = display received data / request json				inbound: N/A | outbound: 777
// layer2 = forming to-be-displayed data / relay request		inbound: 777 | outbound: 888
// layer3 = parse request / communicate with db				inbound: 888 | outbound: DB(3906)
