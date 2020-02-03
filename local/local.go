package main

import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"encoding/json"
	"os"
	"io/ioutil"
)

type Host struct {
	Local	string	`json:"local"`
	Web		string	`json:"web"`
	Was		string	`json:"was"`
}

type Port struct {
	Local	uint16	`json:"local"`
	Web		uint16	`json:"web"`
	Was		uint16	`json:"was"`
}

type Conf struct {
	Host	Host	`json:"host"`
	Port	Port	`json:"port"`
}

type Employee struct {
	ID		uint16
	Name	string
	Start	uint16
}

type Database []Employee

func GetDatabase() Database {
	endpoint := GetEndpoint()
	db, err := sql.Open("mysql", endpoint)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, name, start FROM employee")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer rows.Close()

	everybody := Database{}
	employee := Employee{}
	for rows.Next() {
		err := rows.Scan(&(employee.ID), &(employee.Name), &(employee.Start))
		if err != nil {
			fmt.Println(err)
		}
		everybody = append(everybody, employee)
	}

	return everybody 
}

func PutDatabase(id *uint16, name *string, start *uint16, update Employee) bool {
	endpoint := GetEndpoint()
	db, err := sql.Open("mysql", endpoint)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer db.Close()

	query := "UPDATE employee SET "

	if &(update.ID) != nil {
		query += fmt.Sprintf("id = %d, ", update.ID)
	}
	if &(update.Name) != nil {
		query += fmt.Sprint("name = '%s', ", update.Name)
	}
	if &(update.Start) != nil {
		query += fmt.Sprint("start = %d, ", update.Start)
	}

	query += "WHERE "

	if id != nil {
		query += fmt.Sprintf("id = %d, ", *id)
	}
	if name != nil {
		query += fmt.Sprintf("id = '%s', ", *name)
	}
	if start != nil {
		query += fmt.Sprintf("id = %d, ", *start)
	}

	result, err := db.Exec("UPDATE employee SET id, name, start FROM employee")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}


}

func GetEndpoint() string {
	username, password := os.Getenv("RDS_USERNAME"), os.Getenv("RDS_PASSWORD")

	file, err := os.Open("../conf.json")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer file.Close()


	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	conf := Conf{}
	json.Unmarshal(bytes, &conf)

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/employee", username, password, conf.Host.Web, conf.Port.Web)
} 

func main() {
	
	
	host, port := LoadConf("web")

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/employee", username, password, host, port))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()







}