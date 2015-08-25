package commands

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	. "github.com/leochu/gormeter/summary/stats"
	"github.com/montanaflynn/stats"
)

func GenerateSummary(c *cli.Context) {
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

func getResponseTime(record string) string {
	columns := strings.SplitN(record, ",", 3)
	return columns[1]
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

func check(e error) {
	if e != nil {
		panic(e)
	}
}
