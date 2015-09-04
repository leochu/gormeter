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
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	. "github.com/leochu/gormeter/summary/stats"
	"github.com/montanaflynn/stats"
)

func GenerateSummary(c *cli.Context) {
	inDir, outDir := getPaths(c)
	xml := c.Bool("xml")

	fileInfos, err := ioutil.ReadDir(inDir)
	if err != nil {
		fmt.Printf("Failed to open directory %s: %s\n", inDir, err.Error())
		os.Exit(1)
	}

	if !isDirExist(outDir) {
		err := os.Mkdir(outDir, os.ModePerm)
		check(err)
	}
	outPath := fmt.Sprintf("%ssummary-%d.log", outDir, time.Now().Unix())

	f, err := os.Create(outPath)
	check(err)
	defer f.Close()

	w := bufio.NewWriter(f)

	for _, fileInfo := range fileInfos {
		fileName := fileInfo.Name()

		if !fileInfo.IsDir() && !strings.HasPrefix(fileName, ".") {
			processFile(fileName, inDir, outDir, w, xml)
		}
	}

	contents, _ := ioutil.ReadFile(outPath)
	println(string(contents))
}

func processFile(fileName string, inDir string, outDir string, writer *bufio.Writer, xml bool) {
	filePath := inDir + fileName
	data, err := ioutil.ReadFile(filePath)

	if err != nil {
		fmt.Printf("Failed to open file %s: %s\n", filePath, err.Error())
		os.Exit(1)
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))

	var stats []float64

	for scanner.Scan() {
		record := scanner.Text()

		responseTimeStr, err := getResponseTime(record, xml)
		if err != nil {
			fmt.Println("Could not parse: ", record)
			continue
		}

		responseTime, err := strconv.ParseFloat(responseTimeStr, 64)
		if err != nil {
			fmt.Printf("Failed to parse response time %s: %s\n", responseTimeStr, err.Error())
			os.Exit(1)
		}

		stats = append(stats, responseTime)
	}

	generateSummary(fileName, stats, writer)
	outputStats(outDir, fileName, stats)
}

func outputStats(outDir string, fileName string, stats []float64) {
	outPathLog := fmt.Sprintf("%sresponseTime-%s-%d.log", outDir, fileName, time.Now().Unix())
	f_log, err := os.Create(outPathLog)
	check(err)
	defer f_log.Close()

	w_log := bufio.NewWriter(f_log)

	for _, time := range stats {
		w_log.Write([]byte(fmt.Sprintf("%f\n", time)))
	}
}

func getResponseTime(record string, xml bool) (string, error) {
	if xml {
		re := regexp.MustCompile(" t=\"([0-9]+)\" ")
		match := re.FindStringSubmatch(record)
		if len(match) != 2 {
			return "", errors.New("Could not parse response time")
		}
		return match[1], nil
	}

	// This is used on OSX for CSV
	columns := strings.SplitN(record, ",", 3)
	return columns[1], nil
}

func generateSummary(fileName string, stats []float64, writer *bufio.Writer) {
	summary := getSummary(stats)
	summary.Id = fileName

	buff, _ := json.Marshal(summary)
	fmt.Fprintf(writer, "%s\n", string(buff))

	writer.Flush()
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

func getPaths(c *cli.Context) (string, string) {
	path := c.String("path")
	if path == "" {
		fmt.Println("Please provide --path parameter")
		os.Exit(1)
	}

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	outDir := c.String("out")
	if outDir == "" {
		outDir = path + "summary/"
	}

	if !strings.HasSuffix(outDir, "/") {
		outDir += "/"
	}

	return path, outDir
}

func isDirExist(dir string) bool {
	_, err := os.Stat(dir)
	return !os.IsNotExist(err)
}
