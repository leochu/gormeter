package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/montanaflynn/stats"
)

var flags = []cli.Flag{
	cli.StringFlag{
		Name:  "path",
		Usage: "Path of log files (required)",
	},
	cli.BoolFlag{
		Name:  "json",
		Usage: "Output in json format",
	},
}

var perfFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "path",
		Usage: "Path of log files",
	},
	cli.StringFlag{
		Name:  "httpPath",
		Usage: "Path to summary file of http",
	},
	cli.StringFlag{
		Name:  "httpsPath",
		Usage: "Path to summary file of https",
	},
}

var cliCommands = []cli.Command{
	{
		Name:   "generate",
		Usage:  "generates summary",
		Action: generateSummary,
		Flags:  flags,
	},
	{
		Name:   "perfanalysis",
		Usage:  "Performs analysis on generated summary",
		Action: performAnalysis,
		Flags:  perfFlags,
	},
}

type Summary struct {
	Min                         float64   `json:"min"`
	Max                         float64   `json:"max"`
	Sum                         float64   `json:"sum"`
	Mean                        float64   `json:"mean"`
	Median                      float64   `json:"median"`
	Mode                        []float64 `json:"mode"`
	PopulationVariance          float64   `json:"population_variance"`
	SampleVariance              float64   `json:"sample_variance"`
	StandardDeviationPopulation float64   `json:standard_deviation_population`
	StandardDeviationSample     float64   `json:standard_deviation_sample`
	PercentileOf99              float64   `json:percentile_of_99`
	PercentileNearestRankOf99   float64   `json:percentile_nearest_rank_of_99`
	PercentileOf95              float64   `json:percentile_of_95`
	PercentileNearestRankOf95   float64   `json:percentile_nearest_rank_of_95`
	PercentileOf90              float64   `json:percentile_of_90`
	PercentileNearestRankOf90   float64   `json:percentile_nearest_rank_of_90`
	PercentileOf85              float64   `json:percentile_of_85`
	PercentileNearestRankOf85   float64   `json:percentile_nearest_rank_of_85`
}

func main() {
	fmt.Println()
	app := cli.NewApp()
	app.Name = "jmeter summary"
	app.Commands = cliCommands
	app.CommandNotFound = commandNotFound
	app.Version = "0.1.0"

	app.Run(os.Args)
	os.Exit(0)
}

func generateSummary(c *cli.Context) {
	path := c.String("path")
	if path == "" {
		os.Exit(1)
	}

	jsonFormat := c.Bool("json")

	fileInfos, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Printf("Failed to open directory %s: %s\n", path, err.Error())
		os.Exit(1)
	}

	for _, fileInfo := range fileInfos {
		fileName := fileInfo.Name()

		if !strings.HasPrefix(fileName, ".") {
			if !strings.HasSuffix(path, "/") {
				path += "/"
			}
			processFile(path+fileName, jsonFormat)
		}
	}
}

func performAnalysis(c *cli.Context) {
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

func performAnalysisOnFile(httpSummaryFile, httpsSummaryFile string) {
	httpSummary := unmarshalSummary(httpSummaryFile)
	httpsSummary := unmarshalSummary(httpsSummaryFile)

	changePercent := calculatePercentChange(httpSummary.Mean, httpsSummary.Mean)
	fmt.Printf("The mean response time increased by: %.2f%%\n", changePercent)

	changePercent = calculatePercentChange(httpSummary.Median, httpsSummary.Median)
	fmt.Printf("The median response time increased by: %.2f%%\n", changePercent)

	fmt.Println()
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

func processFile(fileName string, jsonFormat bool) {
	data, err := ioutil.ReadFile(fileName)

	if err != nil {
		fmt.Printf("Failed to open file %s: %s\n", fileName, err.Error())
		os.Exit(1)
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))

	var stats []float64

	for scanner.Scan() {
		record := scanner.Text()

		responseTimeStr := getResponseTime(record)

		responseTime, err := strconv.ParseFloat(responseTimeStr, 64)
		if err != nil {
			fmt.Printf("Failed to parse response time %s: %s\n", responseTimeStr, err.Error())
			os.Exit(1)
		}

		stats = append(stats, responseTime)
	}

	fmt.Printf("Print the summary of file \"%v\":\n", fileName)

	printSummary(fileName, stats, jsonFormat)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getSummary(input []float64) Summary {
	var summary Summary

	summary.Min, _ = stats.Min(input)

	summary.Max, _ = stats.Max(input)
	summary.Sum, _ = stats.Sum(input)
	summary.Mean, _ = stats.Mean(input)
	summary.Median, _ = stats.Median(input)
	summary.Mode, _ = stats.Mode(input)
	summary.PopulationVariance, _ = stats.PopulationVariance(input)
	summary.SampleVariance, _ = stats.SampleVariance(input)
	summary.StandardDeviationPopulation, _ = stats.StandardDeviationPopulation(input)
	summary.StandardDeviationSample, _ = stats.StandardDeviationSample(input)

	// 99%, 95%, 90%, 85%
	percentage := float64(99)
	summary.PercentileOf99, _ = stats.Percentile(input, percentage)
	summary.PercentileNearestRankOf99, _ = stats.PercentileNearestRank(input, percentage)

	percentage = float64(95)
	summary.PercentileOf95, _ = stats.Percentile(input, percentage)
	summary.PercentileNearestRankOf95, _ = stats.PercentileNearestRank(input, percentage)

	percentage = float64(90)
	summary.PercentileOf90, _ = stats.Percentile(input, percentage)
	summary.PercentileNearestRankOf90, _ = stats.PercentileNearestRank(input, percentage)

	percentage = float64(85)
	summary.PercentileOf85, _ = stats.Percentile(input, percentage)
	summary.PercentileNearestRankOf85, _ = stats.PercentileNearestRank(input, percentage)

	return summary
}

func printSummary(fileName string, input []float64, jsonFormat bool) {
	summary := getSummary(input)

	f, err := os.Create(fileName + "summary")
	check(err)

	defer f.Close()
	w := bufio.NewWriter(f)

	if !jsonFormat {
		fmt.Fprintf(w, "Min: %.f\n", summary.Min)                                                // 1.1
		fmt.Fprintf(w, "Max: %.f\n", summary.Max)                                                // 1.1
		fmt.Fprintf(w, "Sum: %.f\n", summary.Sum)                                                // 1.1
		fmt.Fprintf(w, "Mean: %.f\n", summary.Mean)                                              // 1.1
		fmt.Fprintf(w, "Median: %.f\n", summary.Median)                                          // 1.1
		fmt.Fprintf(w, "Mode: %v\n", summary.Mode)                                               // 1.1
		fmt.Fprintf(w, "PopulationVariance: %f\n", summary.PopulationVariance)                   // 1.1
		fmt.Fprintf(w, "SampleVariance: %f\n", summary.SampleVariance)                           // 1.1
		fmt.Fprintf(w, "StandardDeviationPopulation: %f\n", summary.StandardDeviationPopulation) // 1.1
		fmt.Fprintf(w, "StandardDeviationSample: %f\n", summary.StandardDeviationSample)         // 1.1

		fmt.Fprintf(w, "Percentile of 99%%: %v\n", summary.PercentileOf99)                       // 1.1
		fmt.Fprintf(w, "PercentileNearestRank of 99%%: %v\n", summary.PercentileNearestRankOf99) // 1.1

		fmt.Fprintf(w, "Percentile of 95%%: %v\n", summary.PercentileOf95)                       // 1.1
		fmt.Fprintf(w, "PercentileNearestRank of 95%%: %v\n", summary.PercentileNearestRankOf95) // 1.1

		fmt.Fprintf(w, "Percentile of 90%%: %v\n", summary.PercentileOf90)                       // 1.1
		fmt.Fprintf(w, "PercentileNearestRank of 90%%: %v\n", summary.PercentileNearestRankOf90) // 1.1

		fmt.Fprintf(w, "Percentile of 85%%: %v\n", summary.PercentileOf85)                       // 1.1
		fmt.Fprintf(w, "PercentileNearestRank of 85%%: %v\n", summary.PercentileNearestRankOf85) // 1.1
	} else {
		buff, _ := json.Marshal(summary)
		fmt.Fprintf(w, "%s", string(buff))
	}

	w.Flush()

	contents, _ := ioutil.ReadFile(fileName + "summary")
	println(string(contents))
}

func getResponseTime(record string) string {
	columns := strings.SplitN(record, ",", 3)
	return columns[1]
}

func commandNotFound(c *cli.Context, cmd string) {
	fmt.Println("Not a valid command:", cmd)
	os.Exit(1)
}
