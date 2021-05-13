package entities

import (
	"fmt"
	"os"
	"log"
	"bufio"
	"io/ioutil"
	"strings"
	"encoding/json"
)

type Config struct {
	CollectionPath 	string 		`json:"collectionPath"`
	ExactMatch 		bool		`json:"exactMatch"`
	SkippedFiles	[]string 	`json:"skippedFiles"`
	SearchKeywords	[]string 	`json:"searchKeywords"`
}

func LoadConfig(configPath string) Config {
	fmt.Println("Loading config...")

	// read file directly to bytes
	// file, _ := os.Open() -> fbytes := ioutil.ReadAll(file)
	fbytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error reading config file: %s", err)
		os.Exit(1)
	}

	// new config
	var config Config

	// unmarshal into our config struct, return
	json.Unmarshal(fbytes, &config)
	return config
}

func CreateConfig() Config {
	// file doesn't exist: ask for input
	var userCPath string
	var userExactMatch bool
	var userSkippedFiles []string
	var userSearchKeywords []string

	fmt.Print("config.json does not exist: running setup")
	fmt.Print("\nEnter collection path: ")

	reader := bufio.NewReader(os.Stdin)
	inp, _ := reader.ReadString('\n')
	userCPath = strings.TrimSuffix(inp, "\n")

	// read value for exactMatch
	inp = ""
	for {
		fmt.Printf("Require exact keyword match? (y/N): ")
		fmt.Scanf("%s\n", &inp)

		if strings.ToLower(inp) == "y" || strings.ToLower(inp) == "n" {
			break
		}
	}

	if strings.ToLower(inp) == "y" {
		userExactMatch = true
	} else {
		userExactMatch = false
	}

	// read value for skippedFiles

	// read value for keywords
	fmt.Print("Enter keywords to search for (empty + enter to continue)\n")
	for {
		fmt.Print("enter keyword: ")

		// pull input up to \n, remove trailing newline
		inp, _ := reader.ReadString('\n')
		inp = strings.TrimSuffix(inp, "\n")

		// if not empty, read into arr
		if inp != "" {
			userSearchKeywords = append(userSearchKeywords, inp)
		} else {
			break
		}

		inp = ""
	}

	// create config
	config := Config {
		CollectionPath: userCPath,
		ExactMatch: 	userExactMatch,
		SkippedFiles: 	userSkippedFiles,
		SearchKeywords: userSearchKeywords,
	}

	// marshal
	jsonbytes, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		log.Fatalf("Error marshaling json! Err: %s", err)
	}

	fmt.Println("config:", config)
	fmt.Println("jsonbytesstring:", string(jsonbytes))

	// create file
	file, err := os.Create("config.json")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// write, close
	file.Write(jsonbytes)
	file.Close()

	return config
}
