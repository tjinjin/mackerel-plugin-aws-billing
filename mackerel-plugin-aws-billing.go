package main

import (
	"errors"
	"flag"
	"log"
	"time"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/cloudwatch"
	mp "github.com/mackerelio/go-mackerel-plugin-helper"
)

const (
	namespace          = "AWS/Billing"
	region             = "us-east-1"
	metricName         = "EstimatedCharges"
	metricsTypeMaximum = "Maximum"
)

var graphdef = map[string](mp.Graphs){
	"AWS.Billing": mp.Graphs{
		Label: "AWS Billing",
		Unit:  "float",
		Metrics: [](mp.Metrics){
			mp.Metrics{Name: "EstimatedCharges", Label: "EstimatedCharges", Type: "float64"},
		},
	},
}

type metrics struct {
	Name string
	Type string
}

type AwsBillingPlugin struct {
	AccessKeyID     string
	SecretAccessKey string
	CloudWatch      *cloudwatch.CloudWatch
	Name            string
}

// FetchMetrics fetch the metrics
func (p AwsBillingPlugin) FetchMetrics() (map[string]interface{}, error) {
	stat := make(map[string]interface{})

	for _, met := range [...]metrics{
		{Name: metricName, Type: metricsTypeMaximum},
	} {
		v, err := p.getLastPoint(met)
		if err == nil {
			stat[met.Name] = v
		} else {
			log.Printf("%s: %s", met, err)
		}
	}

	return stat, nil
}

func (p *AwsBillingPlugin) prepare() error {
	auth, err := aws.GetAuth(p.AccessKeyID, p.SecretAccessKey, "", time.Now())
	if err != nil {
		return err
	}

	p.CloudWatch, err = cloudwatch.NewCloudWatch(auth, aws.Regions[region].CloudWatchServicepoint)
	if err != nil {
		return err
	}

	return nil
}

func (p AwsBillingPlugin) getLastPoint(metric metrics) (float64, error) {
	now := time.Now()

	dimensions := []cloudwatch.Dimension{
		{
			Name:  "Currency",
			Value: "USD",
		},
	}

	response, err := p.CloudWatch.GetMetricStatistics(&cloudwatch.GetMetricStatisticsRequest{
		Dimensions: dimensions,
		StartTime:  now.Add(time.Duration(21600) * time.Second * -1),
		EndTime:    now,
		MetricName: metric.Name,
		Period:     21600,
		Statistics: []string{metric.Type},
		Namespace:  namespace,
	})
	if err != nil {
		return 0, err
	}
	datapoints := response.GetMetricStatisticsResult.Datapoints
	if len(datapoints) == 0 {
		return 0, errors.New("fetched no datapoints")
	}

	// get a least recently datapoint
	// because a most recently datapoint is not stable.
	least := now
	var latestVal float64
	for _, dp := range datapoints {
		if dp.Timestamp.Before(least) {
			least = dp.Timestamp
			latestVal = dp.Maximum
		}
	}

	return latestVal, nil
}

// GraphDefinition of AwsBillingPlugin
func (p AwsBillingPlugin) GraphDefinition() map[string](mp.Graphs) {
	return graphdef
}

func main() {
	optAccessKeyID := flag.String("access-key-id", "", "AWS Access Key ID")
	optSecretAccessKey := flag.String("secret-access-key", "", "AWS Secret Access Key")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	var plugin AwsBillingPlugin

	plugin.AccessKeyID = *optAccessKeyID
	plugin.SecretAccessKey = *optSecretAccessKey

	err := plugin.prepare()
	if err != nil {
		log.Fatalln(err)
	}

	helper := mp.NewMackerelPlugin(plugin)

	if *optTempfile != "" {
		helper.Tempfile = *optTempfile
	} else {
		helper.Tempfile = "/tmp/mackerel-plugin-aws-billing"
	}

	helper.Run()
}
