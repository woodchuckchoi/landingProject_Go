package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
	"path"
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
	endPoint 	string 	= "http://" + LoadConf("web")
	target   	string 	= "http://" + LoadConf("was")
	client				= &http.Client{}
	funcList			= []func(){ListEmployee, SearchEmployee, CreateEmployee, EditEmployee, DeleteEmployee}
)

var (
	getAllEndpoint string = target + "/employee"
	newEndpoint    string = getAllEndpoint + "/new"
	idEndpoint     string = getAllEndpoint + "/id"
	nameEndpoint   string = getAllEndpoint + "/name"
)

func UpperDir(original string) string {
	lastSlash := 0
	for i, val := range original {
		if val == '/' {
			lastSlash = i
		}
	}
	return original[:lastSlash]
}

func LoadConf(target string) string {
	curFile, _ := os.Executable()
	curDir := path.Dir(curFile)
	
	f, err := os.Open(UpperDir(curDir) + "/conf.json")
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

func ClearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func CreateQuery() int {
	var num int

	for {
		ClearScreen()

		fmt.Println("0. LIST ALL EMPLOYEES")
		fmt.Println("1. SEARCH EMPLOYEE")
		fmt.Println("2. CREATE EMPLOYEE")
		fmt.Println("3. EDIT EMPLOYEE")
		fmt.Println("4. DELETE EMPLOYEE")
		fmt.Println("5. QUIT")
		fmt.Printf("CHOOSE AN OPTION: ")

		_, err := fmt.Scanf("%d", &num)

		if err != nil || num < 0 || num > 5 {
			fmt.Println()
			fmt.Println("INVALID INPUT")
			time.Sleep(time.Second * 2)
			continue
		}
		break
	}
	return num
}

func SearchEmployee() {
	ClearScreen()
	var dest string

	var field string
	var value string
	for {
		fmt.Printf("(i)d/(n)ame: ")
		if fmt.Scanf("%s", &field); field == "i" || field == "n" {
			break
		}
	}

	var query string
	switch field {
	case "i":
		query = "ID: "
		dest = idEndpoint
	case "n":
		query = "NAME: "
		dest = nameEndpoint
	}

	fmt.Printf(query)
	fmt.Scanf("%s", &value)

	dest += fmt.Sprintf("/%s", value)
	
	message := bytes.NewBufferString(fmt.Sprintf(jsonData, dest, ""))
	
	req, err := http.NewRequest(http.MethodGet, endPoint, message)
	if err != nil {
		log.Fatal(err)
	}
	
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println(string(bytes))
	var next string
	fmt.Scanf("%s", &next)
}

func ListEmployee() {
	ClearScreen()
	dest := getAllEndpoint
	
	message := bytes.NewBufferString(fmt.Sprintf(jsonData, dest, ""))
	
	req, err := http.NewRequest(http.MethodGet, endPoint, message)
	if err != nil {
		log.Fatal(err)
	}
	
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println(string(bytes))
	var next string
	fmt.Scanf("%s", &next)
}

func CreateEmployee() {
	ClearScreen()
	var employee Person

	fmt.Printf("ID: ")
	fmt.Scanf("%d", &(employee.ID))
	fmt.Printf("NAME: ")
	fmt.Scanf("%s", &(employee.Name))
	fmt.Printf("SALARY: ")
	fmt.Scanf("%d", &(employee.Salary))

	bytesData, err := json.Marshal(employee)
	if err != nil {
		log.Fatal(err)
	}

	message := fmt.Sprintf(jsonData, newEndpoint, string(bytesData))

	req, err := http.NewRequest(http.MethodPost, endPoint, bytes.NewBufferString(message))
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	
	printData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(printData))

	var next string
	fmt.Scanf("%s", &next)
}

func EditEmployee() {
	ClearScreen()
	
	var choice string
	for {
		fmt.Printf("(i)d/(n)ame: ")
		
		_, err := fmt.Scanf("%s", &choice)
		if err != nil || (choice != "i" && choice != "n"){
			fmt.Println("INVALID INPUT")
			continue
		}
		break
	}

	var dest string
	var value string

	if choice == "i" {
		fmt.Printf("ID: ")
		dest = idEndpoint
	} else {
		fmt.Printf("NAME: ")
		dest = nameEndpoint
	}

	fmt.Scanf("%s", &value)
	dest += fmt.Sprintf("/%s", value)

	var employee Person

	fmt.Println("CHANGES TO BE MADE")
	fmt.Printf("ID: ")
	fmt.Scanf("%d", &(employee.ID))
	fmt.Printf("NAME: ")
	fmt.Scanf("%s", &(employee.Name))
	fmt.Printf("SALARY: ")
	fmt.Scanf("%d", &(employee.Salary))

	bytesData, err := json.Marshal(employee)
	if err != nil {
		log.Fatal(err)
	}
	
	message := fmt.Sprintf(jsonData, dest, string(bytesData))

	req, err := http.NewRequest(http.MethodPut, endPoint, bytes.NewBufferString(message))
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	printData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(printData))

	var next string
	fmt.Scanf("%s", &next)
}

func DeleteEmployee() {
	ClearScreen()

	var choice string
	for {
		fmt.Printf("(i)d/(n)ame: ")
		
		_, err := fmt.Scanf("%s", &choice)
		if err != nil || (choice != "i" && choice != "n"){
			fmt.Println("INVALID INPUT")
			continue
		}
		break
	}

	var dest string
	var value string

	if choice == "i" {
		fmt.Printf("ID: ")
		dest = idEndpoint
	} else {
		fmt.Printf("NAME: ")
		dest = nameEndpoint
	}

	fmt.Scanf("%s", &value)
	dest += fmt.Sprintf("/%s", value)

	message := fmt.Sprintf(jsonData, dest, "")

	req, err := http.NewRequest(http.MethodDelete, endPoint, bytes.NewBufferString(message))
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	printData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(printData))

	var next string
	fmt.Scanf("%s", &next)
}

func main() {

	for {
		choice := CreateQuery()
		if choice == 5 {
			return
		}else{
			funcList[choice]()
		}
	}
	
}
