package utils

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type T struct {
	Service struct {
		Name  string
		Fault string
		Value string
	}
}

var data = `
service:
 name: auth-service
 fault: latency
 value: 2000
`

func ParseYaml(filepath string) {
	t := T{}
	data, err := os.ReadFile(filepath)
	
	if err!= nil{
		log.Fatalf("error: %v", err)
	}

	err = yaml.Unmarshal([]byte(data), &t)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t:\n%+v\n\n", t)
	fmt.Println(t.Service.Value)
}
