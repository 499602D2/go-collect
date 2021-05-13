/*
Go-collect takes the path to a folder comprised of gzipped
files as input, and attempts to parse all found archives and
their contents against user-specified search criteria.

First version created on 2021.05.08
*/

package main

import (
	"os"
	"io"
	"fmt"
	"log"
	"bufio"
	"strings"
	"time"
	"path/filepath"
	"io/ioutil"
	"archive/tar"
	"compress/gzip"
	"github.com/dustin/go-humanize"
	"main/entities"
)

func search(config entities.Config, archivepath string) ([]string, int64) {
	// join paths, open file
	file, err := os.Open(filepath.Join(config.CollectionPath, archivepath))
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	// open file as a gzip file
	gzfile, err := gzip.NewReader(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// read contents
	tarReader := tar.NewReader(gzfile)

	// store matches in a slice
	var matches []string
	var parsed int64 = 0

	for {
		// get next file entry
		_, err := tarReader.Next()
		if err == io.EOF {
			fmt.Println("\t↳ end of archive")
			break
		}
		
		if err != nil {
			log.Fatal(err)
		}

		// create a Scanner for reading line by line
		scanner := bufio.NewScanner(tarReader)

		// if scanner runs out of buffer, throw more bytes at it
		if scanner.Err() == bufio.ErrTooLong {
			fmt.Println("Attempting to avoid bufio.ErrTooLong...")
			scanner := bufio.NewScanner(file)
			buf := make([]byte, 0, 64*1024)
			scanner.Buffer(buf, 1024*1024)
		}

		// line reading loop
		var lnum int64 = 0
		for scanner.Scan() {
			// read the current last read line of text
			line := scanner.Text()

			// check for a match
			var splits []string
			if strings.Contains(line, ":") == true {
				splits = strings.Split(line, ":")
			} else if strings.Contains(line, ";") == true {
				splits = strings.Split(line, ";")
			} else if strings.Contains(line, "|") == true {
				splits = strings.Split(line, "|")
			} else {
				splits = append(splits, line, line)
			}

			// check for a match
			for _, target := range config.SearchKeywords {
				if config.ExactMatch == true {
					if splits[0] == target {
						fmt.Printf("\t↣ %s\n", line)
						matches = append(matches, splits[0] + ":" + splits[1])
					}
				} else {
					if strings.Contains(splits[0], target) == true {
						fmt.Printf("\t↣ %s\n", line)
						matches = append(matches, splits[0] + ":" + splits[1])
					}
				}
			}

			lnum++
		}

		parsed += lnum

		// check for error at this point
		if scanner.Err() != nil && scanner.Err() != bufio.ErrTooLong {
			fmt.Println("scanner raised error: ", scanner.Err())
			os.Exit(1)
		}
	}

	// return all matches
	return matches, parsed
}


func loadConfig(configfPath string) entities.Config {
	// if config doesn't exist, create the file
	if _, err := os.Stat(configfPath); os.IsNotExist(err) {
		return entities.CreateConfig()
	}

	// file exists: load
	return entities.LoadConfig(configfPath)
}


func main() {
	const vnum = "2021.05.13"
	fmt.Printf("Go-collect %s started at %s\n\n",
		vnum, time.Now().Format(time.Stamp))

	// log start time
	tstart := time.Now().Unix()

	// load config
	config := loadConfig("config.json")
	fmt.Println("config:", config)

	// list all files in collection path
	files, err := ioutil.ReadDir(config.CollectionPath)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// save all matches in an array
	allMatches := make([]string, 0)

	// keep track of statistics
	// humanize.Comma wants an int64, hence the type
	var parseCount int64 = 0

	// iterate over all files in target directory
	for _, file := range files {
		if file.Name()[0] != '.' {
			// skip files we've already parsed
			fileParsed := false
			for _, fname := range config.SkippedFiles {
				if fname == file.Name() {
					fileParsed = true
				}
			}

			if fileParsed == true {
				fmt.Println("⃕ Skipping", file.Name())
				continue
			}

			fmt.Println("Opening archive:", file.Name())

			// parse all files
			hits, parsed := search(config, file.Name())
			parseCount += parsed

			// extend allMatches to fit hits
			temp := make([]string, len(allMatches) + len(hits))
			copy(temp, allMatches)
			copy(temp, hits)

			// rename
			allMatches = temp
		}
	}

	// clean matches by removing duplicates
	var cleanedMatches []string
	for _, match := range allMatches {
		exists := false
		for _, cmatch := range cleanedMatches {
			if match == cmatch {
				exists = true
				break
			}
		}

		if exists == false {
			cleanedMatches = append(cleanedMatches, match)
		}
	}

	// log average processing speed
	tend := time.Now().Unix()
	eps := parseCount / (tend - tstart)

	// print all hits
	fmt.Printf(
		"\n→ Search completed in %d seconds: found %d hit(s) out of %s\n",
		tend - tstart,
		len(cleanedMatches),
		humanize.Comma(parseCount))

	fmt.Printf(
		"→ Average processing speed: %s entries per second\n\n",
		humanize.Comma(eps))

	for i, hit := range cleanedMatches {
		fmt.Printf("hit %5d: %s\n", i, hit)
	}

	if len(cleanedMatches) > 0 {
		// create output file
		fname := fmt.Sprintf("matches-%d.txt", time.Now().Unix())
		file, err := os.Create(filepath.Join("output", fname))
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		// dump hits to file
		writer := bufio.NewWriter(file)
		for _, line := range cleanedMatches {
			_, err := writer.WriteString(line + "\n")
			if err != nil {
				log.Fatalf("Error writing to file: %s", err.Error())
			}
		}

		writer.Flush()
	}
}
