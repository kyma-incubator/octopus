package main

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/olekukonko/tablewriter"
	"io"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"strings"
	"time"
	// or
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // because of problems with vendor, currently only GCP is supported
)

func main() {
	if len(os.Args) != 2 {
		panic("one argument required: suite name")
	}
	suiteName := os.Args[1]
	if err := v1alpha1.SchemeBuilder.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}

	cl, err := client.New(config.GetConfigOrDie(), client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		panic(err)
	}

	cts := &v1alpha1.ClusterTestSuite{}

	err = cl.Get(context.Background(), types.NamespacedName{Name: suiteName}, cts)
	if err != nil {
		panic(fmt.Sprintf("failed to get ClusterTestSuite %s: got error: %v", suiteName, err))

	}
	printer := &TablePrinter{}
	printer.Print(*cts, os.Stdout)
}

type TablePrinter struct {
}

func (tp *TablePrinter) Print(suite v1alpha1.ClusterTestSuite, out io.Writer) error {
	data := New(suite)
	rows := [][]string{
		{"Name", data.Name},
		{"Concurrency", fmt.Sprintf("%d", data.Concurrency)},
	}
	if data.MaxRetries > 0 {
		rows = append(rows, []string{"Max Retires", fmt.Sprintf("%d", data.MaxRetries)})
	} else {
		rows = append(rows, []string{"Count", fmt.Sprintf("%d", data.Count)})
	}
	rows = append(rows, []string{"Duration", fmt.Sprintf("%v", data.Duration)})
	rows = append(rows, []string{"Condition", data.Condition})
	rows = append(rows, []string{"Tests", fmt.Sprintf("%d", data.TestNumber)})
	rows = append(rows, []string{"In Progress", fmt.Sprintf("%d", data.InProgressTestsNumber)})
	rows = append(rows, []string{"Success", fmt.Sprintf("%d", data.SuccessfulTestsNumber)})
	rows = append(rows, []string{"Failures", fmt.Sprintf("%d", data.FailedTestsNumber)})
	rows = append(rows, []string{"Executions.", fmt.Sprintf("%d", data.ExecutionNumber)})
	rows = append(rows, []string{"Failed tests", data.FailedTestNames})

	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Name", "Value"})

	for _, r := range rows {
		table.Append(r)
	}
	table.Render() // Send output
	return nil
}

type SuiteData struct {
	Name                  string
	Concurrency           int
	Count                 int
	MaxRetries            int
	StartTime             *time.Time
	CompletionTime        *time.Time
	Duration              time.Duration
	Condition             string
	TestNumber            int
	ExecutionNumber       int
	InProgressTestsNumber int
	SuccessfulTestsNumber int
	FailedTestsNumber     int
	FailedTestNames       string
}

func New(in v1alpha1.ClusterTestSuite) SuiteData {
	var startTime *time.Time
	var completionTime *time.Time
	if in.Status.StartTime != nil {
		startTime = &in.Status.StartTime.Time
	}
	if in.Status.CompletionTime != nil {
		completionTime = &in.Status.CompletionTime.Time
	}
	var duration time.Duration
	if startTime != nil {
		if completionTime == nil {
			duration = time.Since(*startTime)
		} else {
			duration = completionTime.Sub(*startTime)
		}

	}
	var conditions []string
	for _, cond := range in.Status.Conditions {
		if cond.Status == v1alpha1.StatusTrue {
			conditions = append(conditions, string(cond.Type))
		}
	}

	var execNumber int
	for _, tr := range in.Status.Results {
		execNumber += len(tr.Executions)
	}

	var inProgress int
	var success int
	var failed int
	var failedNames []string

	for _, tr := range in.Status.Results {
		switch tr.Status {
		case v1alpha1.TestNotYetScheduled:
			inProgress++
		case v1alpha1.TestScheduled:
			inProgress++
		case v1alpha1.TestRunning:
			inProgress++
		case v1alpha1.TestUnknown:
			inProgress++
		case v1alpha1.TestFailed:
			failed++
			failedNames = append(failedNames, tr.Name)
		case v1alpha1.TestSucceeded:
			success++
		case v1alpha1.TestSkipped:
			success++
		}
	}

	failedTests := "-"
	if len(failedNames) > 0 {
		failedTests = strings.Join(failedNames, ",")
	}

	out := SuiteData{
		Name:                  in.Name,
		Concurrency:           int(in.Spec.Concurrency),
		Count:                 int(in.Spec.Count),
		MaxRetries:            int(in.Spec.MaxRetries),
		StartTime:             startTime,
		CompletionTime:        completionTime,
		Duration:              duration,
		Condition:             strings.Join(conditions, ","),
		TestNumber:            len(in.Status.Results),
		ExecutionNumber:       execNumber,
		InProgressTestsNumber: inProgress,
		SuccessfulTestsNumber: success,
		FailedTestsNumber:     failed,
		FailedTestNames:       failedTests,
	}
	return out
}
