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

	"github.com/codegangsta/cli"
	. "github.com/leochu/gormeter/summary/stats"
)

func PerformAnalysis(c *cli.Context) {
	path := c.String("path")
	if path == "" {
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
		fileName := fileInfo.Name()

		data, err := ioutil.ReadFile(path + fileName)
		if err != nil {
			fmt.Printf("Failed to open file %s: %s\n", fileName, err.Error())
			os.Exit(1)
		}

		scanner := bufio.NewScanner(bytes.NewReader(data))

		for scanner.Scan() {
			record := scanner.Bytes()

			// fmt.Println(string(record))
			var summary Summary
			err = json.Unmarshal(record, &summary)
			if err != nil {
				fmt.Printf("Failed to unmarshal data in file %s: %s\n", fileName, err.Error())
				os.Exit(1)
			}

			summaryMap[summary.Id] = summary
		}
	}

	for fileName, summary := range summaryMap {
		if isHttpsFile(fileName) {
			httpSummary, err := getHTTPTestSummary(fileName, summaryMap)
			if err != nil {
				fmt.Printf("Couldn't find HTTP file for %s\n", fileName)
				continue
			}

			fmt.Printf("Performing analysis on %s\n", fileName)
			performAnalysis(httpSummary, summary)
		}
	}
}

func isHttpsFile(fileName string) bool {
	re, err := regexp.Compile(".*https.*log")
	if err != nil {
		fmt.Printf("Error:%s\n", err.Error())
		os.Exit(1)
	}

	return re.Match([]byte(fileName))
}

func performAnalysis(httpSummary, httpsSummary Summary) {
	changePercent := calculatePercentChange(httpSummary.Mean, httpsSummary.Mean)
	fmt.Printf("The mean response time increased by: %.2f%%\n", changePercent)

	changePercent = calculatePercentChange(httpSummary.Median, httpsSummary.Median)
	fmt.Printf("The median response time increased by: %.2f%%\n\n", changePercent)
}

func getHTTPTestSummary(httpsTest string, summaryMap map[string]Summary) (Summary, error) {
	httpTest := strings.Replace(httpsTest, "https", "http", 1)

	fileNameElements := strings.Split(httpTest, "-")
	len := len(fileNameElements) - 1
	httpFileNameExp := strings.Join(fileNameElements[:len], "-") + ".*log"
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
