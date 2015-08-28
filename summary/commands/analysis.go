package commands

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	. "github.com/leochu/gormeter/summary/stats"
)

func PerformAnalysis(c *cli.Context) {
	path := c.String("path")
	if path == "" {
		fmt.Println("Please use --path to provide the path of log")
		os.Exit(1)
	}

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	fileInfos, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Printf("Failed to open directory %s: %s\n", path, err.Error())
		os.Exit(1)
	}

	summaryMap := make(map[string]Summary)
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue
		}

		fileName := fileInfo.Name()

		data, err := ioutil.ReadFile(path + fileName)
		if err != nil {
			fmt.Printf("Failed to open file %s: %s\n", fileName, err.Error())
			os.Exit(1)
		}

		scanner := bufio.NewScanner(bytes.NewReader(data))

		for scanner.Scan() {
			record := scanner.Bytes()

			var summary Summary
			err = json.Unmarshal(record, &summary)
			if err != nil {
				fmt.Printf("Failed to unmarshal data in file %s: %s\n", fileName, err.Error())
				os.Exit(1)
			}

			summaryMap[summary.Id] = summary
		}
	}

	// Create the output path if not exit.
	outDir := path + "analysis/"

	if !isDirExist(outDir) {
		fmt.Println("Dir not exit.")
		err := os.Mkdir(outDir, os.ModePerm)
		check(err)
	}

	outPath := fmt.Sprintf("%sanalysis-%d.log", outDir, time.Now().Unix())
	f, err := os.Create(outPath)
	check(err)

	fmt.Printf("Create output file: %s\n\n", outPath)

	defer f.Close()

	w := bufio.NewWriter(f)

	for fileName, summary := range summaryMap {
		if isHttpsFile(fileName) {
			httpSummary, err := getHTTPTestSummary(fileName, summaryMap)
			if err != nil {
				fmt.Printf("Couldn't find HTTP file for %s\n", fileName)
				continue
			}

			performAnalysis(fileName, httpSummary, summary, w)
		}
	}

	contents, _ := ioutil.ReadFile(outPath)
	println(string(contents))
}

func isHttpsFile(fileName string) bool {
	re, err := regexp.Compile(".*https.*log")
	if err != nil {
		fmt.Printf("Error:%s\n", err.Error())
		os.Exit(1)
	}

	return re.Match([]byte(fileName))
}

func performAnalysis(fileName string, httpSummary, httpsSummary Summary, writer *bufio.Writer) {
	fmt.Fprintf(writer, "Performing analysis on %s\n", fileName)

	changePercent := calculatePercentChange(httpSummary.Mean, httpsSummary.Mean)
	fmt.Fprintf(writer, "The mean response time increased by: %.2f%% (From %.2f to %.2f)\n", changePercent, httpSummary.Mean, httpsSummary.Mean)

	changePercent = calculatePercentChange(httpSummary.Median, httpsSummary.Median)
	fmt.Fprintf(writer, "The median response time increased by: %.2f%% (From %v to %v)\n\n", changePercent, httpSummary.Median, httpsSummary.Median)

	writer.Flush()
}

func getHTTPTestSummary(httpsTest string, summaryMap map[string]Summary) (Summary, error) {
	httpTest := strings.Replace(httpsTest, "https", "http", 1)

	sep := "_"

	fileNameElements := strings.Split(httpTest, sep)
	len := len(fileNameElements) - 1
	httpFileNameExp := strings.Join(fileNameElements[:len], sep) + ".*log"

	re, err := regexp.Compile(httpFileNameExp)
	if err != nil {
		fmt.Printf("Error in compiling regexp:%s\n", err.Error())
		os.Exit(1)
	}

	for fileName, summary := range summaryMap {
		if re.Match([]byte(fileName)) {
			return summary, nil
		}
	}

	return Summary{}, errors.New("http file not found")
}

func calculatePercentChange(value1, value2 float64) float64 {
	return 100 * ((value2 - value1) / value1)
}
