package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

var promPath string
var promConfig = filepath.Join("..", "..", "documentation", "examples", "prometheus.yml")

var _ = func() bool {
	testing.Init()
	return true
}()

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		os.Exit(m.Run())
	}
	var err error
	promPath, err = os.Getwd()
	if err != nil {
		fmt.Printf("can't get current dir :%s \n", err)
		os.Exit(1)
	}
	promPath = filepath.Join(promPath, "testData/blm_prometheus.log")
	fmt.Println(promPath)
	//build := exec.Command("go", "build", "-o", promPath)
	//output, err := build.CombinedOutput()
	//if err != nil {
	//	fmt.Printf("compilation error :%s \n", output)
	//	os.Exit(1)
	//}

	exitCode := m.Run()
	//os.Remove(promPath)
	//os.RemoveAll(promData)
	os.Exit(exitCode)
}

func TestSerializationfs(t *testing.T) {
	var req prompb.WriteRequest
	var ts []*prompb.TimeSeries
	var tse prompb.TimeSeries
	var sample prompb.Sample
	var labels []*prompb.Label
	labelValue1 := prompb.Label{Name: "__name__", Value: "testLabel"}
	labelValue2 := prompb.Label{Name: "instance", Value: "testTagInstance"}
	labels = append(labels, &labelValue1, &labelValue2)

	testfile, err := os.OpenFile(promPath, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("Open file error!", err)
		return
	}
	defer testfile.Close()
	fmt.Println(promPath)
	buf := bufio.NewReader(testfile)
	i := 0
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("File read ok! line:", i)
				break
			} else {
				fmt.Println("Read file error!", err)
				return
			}
		}
		if strings.Contains(line, "server.go:201:") {
			sa := strings.Split(line, " ")
			sample.Timestamp, _ = strconv.ParseInt(sa[7][:(len(sa[7])-1)], 10, 64)
			sample.Value, _ = strconv.ParseFloat(sa[9][:(len(sa[9])-1)], 64)
			tse.Samples = append(tse.Samples, sample)
			tse.Labels = labels
			ts = append(ts, &tse)
			req.Timeseries = ts
			fmt.Print(ts)
		}
	}
	_, writer := buildClients()
	ProcessReq(req, writer)
}

func TestRead(t *testing.T) {
	var req prompb.ReadRequest
	var query prompb.Query
	var nameLabel = prompb.LabelMatcher{
		Type: prompb.LabelMatcher_EQ,
		Name: model.MetricNameLabel,
		// inner metric
		Value: "go_goroutines",
	}
	var matchers = prompb.LabelMatcher{
		Type:  prompb.LabelMatcher_EQ,
		Name:  "job",
		Value: "prometheus",
	}

	query.StartTimestampMs = 1618085130781
	query.EndTimestampMs = 1619863860781
	query.Matchers = append(query.Matchers, &matchers, &nameLabel)
	req.Queries = append(req.Queries, &query)

	reader, _ := buildClients()

	reader.Read(&req)

}

func TestTimeData(t *testing.T) {
	var timeStampValue = 1619870400000
	fmt.Println(toTimestamp(int64(timeStampValue)).Format(time.RFC3339))
}

func toTimestamp(milliseconds int64) time.Time {
	sec := milliseconds / 1000
	nsec := (milliseconds - (sec * 1000)) * 1000000
	return time.Unix(sec, nsec)
}
