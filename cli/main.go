package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gitlab.com/pagalguy/loki/ds_to_sql"
)

func Walker(done <-chan bool, root string) (<-chan string, <-chan error) {

	paths := make(chan string)
	errc := make(chan error, 1)

	go func() {

		defer close(paths)

		errc <- filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

			if err != nil {
				return err
			}

			if !info.Mode().IsRegular() {
				return nil
			}

			baseFilename := filepath.Base(path)

			if !strings.HasPrefix(baseFilename, "output") {
				return nil
			}

			select {
			case paths <- path:
			case <-done:
				return errors.New("Walk done")
			}

			return nil

		})

	}()

	return paths, errc
}

func Reader(done <-chan bool, entityName string, paths <-chan string, csvChan chan<- [][]string) {

	for path := range paths {
		log.Printf("Reading file: %s\n", path)
		select {
		case csvChan <- ConvertDSToCSV(path, entityName):
		case <-done:
			return
		}
	}
}

func ConvertDSToCSV(dsFilePath string, dsEntityName string) [][]string {

	newClassCreator := func() ds_to_sql.CSVMixin {

		if dsEntityName == "Entity" {
			return &Entity{}
		} else if dsEntityName == "Follow" {
			return &Follow{}
		} else {
			return &Blank{}
		}
	}

	csvRows, err := ds_to_sql.ReadDSFile(dsFilePath, newClassCreator)

	if err != nil {
		return [][]string{
			[]string{},
		}
	}

	return *csvRows
}

func Writer(csvOutputFolder string, csvChan <-chan [][]string) {

	for csvRows := range csvChan {

		for edgeName, edgeRows := range GroupOutputRows(csvRows) {

			csvFilename := csvOutputFolder + "/" + edgeName + ".csv"
			csvFile, _ := os.OpenFile(csvFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

			csvWriter := csv.NewWriter(csvFile)

			csvWriter.WriteAll(edgeRows)

			csvWriter.Flush()
			csvFile.Close()
		}
	}
}
func Runner(dsBackupsFolder string, csvOutputFolder string, entityName string) error {

	done := make(chan bool)
	defer close(done)

	paths, errc := Walker(done, dsBackupsFolder)

	csvChan := make(chan [][]string)
	var wg sync.WaitGroup

	const workers = 20
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(i int) {
			Reader(done, entityName, paths, csvChan)
			wg.Done()
		}(i)
	}

	go func() {
		wg.Wait()
		close(csvChan)
	}()

	Writer(csvOutputFolder, csvChan)

	// Check whether the Walk failed.
	if err := <-errc; err != nil { // HLerrc
		return err
	}

	return nil
}

func GroupOutputRows(outputRows [][]string) map[string][][]string {

	groupedMap := make(map[string][][]string)

	for _, row := range outputRows {

		if _, ok := groupedMap[row[0]]; !ok {
			groupedMap[row[0]] = make([][]string, 0)
		}

		groupedMap[row[0]] = append(groupedMap[row[0]], row[1:])
	}

	return groupedMap
}

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {

	flag.Parse()

	dsBackupsFolder := flag.Arg(0)
	csvOutputFolder := flag.Arg(1)
	entityName := flag.Arg(2)

	_ = os.MkdirAll(csvOutputFolder, os.ModePerm)

	_ = RemoveContents(csvOutputFolder)

	err := Runner(dsBackupsFolder, csvOutputFolder, entityName)

	if err != nil {
		log.Fatal(err)
	}
}

// https://blog.golang.org/pipelines/bounded.go
// https://blog.golang.org/pipelines
