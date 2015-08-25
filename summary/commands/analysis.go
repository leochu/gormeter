package commands

import (
	"encoding/json"
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
		httpSummaryFile := c.String("httpPath")
		if httpSummaryFile == "" {
			fmt.Printf("Required option --httpPath not provided\n")
			os.Exit(1)
		}

		httpsSummaryFile := c.String("httpsPath")
		if httpsSummaryFile == "" {
			fmt.Printf("Required option --httpsPath not provided\n")
			os.Exit(1)
		}
		performAnalysisOnFile(httpSummaryFile, httpsSummaryFile)
	} else {
		fileInfos, err := ioutil.ReadDir(path)
		if err != nil {
			fmt.Printf("Failed to open directory %s: %s\n", path, err.Error())
			os.Exit(1)
		}
		re, err := regexp.Compile(".*https.*logsummary")
		if err != nil {
			fmt.Printf("Error:%s\n", err.Error())
			os.Exit(1)
		}

		fileNameMap := make(map[string]string)
		for _, fileInfo := range fileInfos {
			fileName := fileInfo.Name()
			if re.Match([]byte(fileName)) {
				httpFileName := getMatchingFileName(fileName, fileInfos)
				if httpFileName == "" {
					fmt.Printf("No matching file for %s\n", fileName)
				} else {
					fileNameMap[fileName] = httpFileName
				}
			}
		}

		for httpsSummaryFile, httpSummaryFile := range fileNameMap {
			fmt.Printf("Performing analysis on %s and %s\n", httpsSummaryFile, httpSummaryFile)
			performAnalysisOnFile(path+"/"+httpSummaryFile, path+"/"+httpsSummaryFile)
		}
	}
}

func performAnalysisOnFile(httpSummaryFile, httpsSummaryFile string) {
	httpSummary := unmarshalSummary(httpSummaryFile)
	httpsSummary := unmarshalSummary(httpsSummaryFile)

	changePercent := calculatePercentChange(httpSummary.Mean, httpsSummary.Mean)
	fmt.Printf("The mean response time increased by: %.2f%%\n", changePercent)

	changePercent = calculatePercentChange(httpSummary.Median, httpsSummary.Median)
	fmt.Printf("The median response time increased by: %.2f%%\n", changePercent)

	fmt.Println()
}

func getMatchingFileName(fileName string, fileInfos []os.FileInfo) string {
	tmpFileName := strings.Replace(fileName, "https", "http", 1)
	fileNameElements := strings.Split(tmpFileName, "-")
	len := len(fileNameElements) - 1
	httpFileNameExp := strings.Join(fileNameElements[:len], "-") + ".*logsummary"
	re, err := regexp.Compile(httpFileNameExp)
	if err != nil {
		fmt.Printf("Error in compiling regexp:%s\n", err.Error())
		os.Exit(1)
	}
	var httpFileName string
	for _, fileInfo := range fileInfos {
		name := fileInfo.Name()
		if re.Match([]byte(name)) {
			httpFileName = name
			break
		}
	}
	return httpFileName
}

func calculatePercentChange(value1, value2 float64) float64 {
	return 100 * ((value2 - value1) / value1)
}

func unmarshalSummary(fileName string) Summary {
	var summary Summary
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Printf("Failed to open file %s: %s\n", fileName, err.Error())
		os.Exit(1)
	}

	err = json.Unmarshal(data, &summary)
	if err != nil {
		fmt.Printf("Failed to unmarshal data in file %s: %s\n", fileName, err.Error())
		os.Exit(1)
	}

	return summary
}
