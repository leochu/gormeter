package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
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
}

var cliCommands = []cli.Command{
	{
		Name:   "generate",
		Usage:  "generates summary",
		Action: generateSummary,
		Flags:  flags,
	},
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
			processFile(path + fileName)
		}

	}

}

func processFile(fileName string) {
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

	printSummary(fileName, stats)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func printSummary(fileName string, input []float64) {
	f, err := os.Create(fileName + "summary")
	check(err)

	defer f.Close()

	w := bufio.NewWriter(f)

	a, _ := stats.Min(input)
	fmt.Fprintf(w, "Min: %.f\n", a) // 1.1
	a, _ = stats.Max(input)
	fmt.Fprintf(w, "Max: %.f\n", a) // 1.1
	a, _ = stats.Sum(input)
	fmt.Fprintf(w, "Sum: %.f\n", a) // 1.1
	a, _ = stats.Mean(input)
	fmt.Fprintf(w, "Mean: %.f\n", a) // 1.1
	a, _ = stats.Median(input)
	fmt.Fprintf(w, "Median: %.f\n", a) // 1.1
	m, _ := stats.Mode(input)
	fmt.Fprintf(w, "Mode: %v\n", m) // 1.1
	a, _ = stats.PopulationVariance(input)
	fmt.Fprintf(w, "PopulationVariance: %f\n", a) // 1.1
	a, _ = stats.SampleVariance(input)
	fmt.Fprintf(w, "SampleVariance: %f\n", a) // 1.1
	a, _ = stats.StandardDeviationPopulation(input)
	fmt.Fprintf(w, "StandardDeviationPopulation: %f\n", a) // 1.1
	a, _ = stats.StandardDeviationSample(input)
	fmt.Fprintf(w, "StandardDeviationSample: %f\n", a) // 1.1

	// 99%, 95%, 90%, 85%
	percentage := float64(99)
	a, _ = stats.Percentile(input, percentage)
	fmt.Fprintf(w, "Percentile of %v%%: %v\n", percentage, a) // 1.1

	a, _ = stats.PercentileNearestRank(input, percentage)
	fmt.Fprintf(w, "PercentileNearestRank of %v%%: %v\n", percentage, a) // 1.1

	percentage = float64(95)
	a, _ = stats.Percentile(input, percentage)
	fmt.Fprintf(w, "Percentile of %v%%: %v\n", percentage, a) // 1.1

	a, _ = stats.PercentileNearestRank(input, percentage)
	fmt.Fprintf(w, "PercentileNearestRank of %v%%: %v\n", percentage, a) // 1.1

	percentage = float64(90)
	a, _ = stats.Percentile(input, percentage)
	fmt.Fprintf(w, "Percentile of %v%%: %v\n", percentage, a) // 1.1

	a, _ = stats.PercentileNearestRank(input, percentage)
	fmt.Fprintf(w, "PercentileNearestRank of %v%%: %v\n", percentage, a) // 1.1

	percentage = float64(85)
	a, _ = stats.Percentile(input, percentage)
	fmt.Fprintf(w, "Percentile of %v%%: %v\n", percentage, a) // 1.1

	a, _ = stats.PercentileNearestRank(input, percentage)
	fmt.Fprintf(w, "PercentileNearestRank of %v%%: %v\n", percentage, a) // 1.1

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
