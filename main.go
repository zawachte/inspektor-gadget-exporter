package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/blang/semver"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zawachte/inspektor-gadget-exporter/collector"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/inspektor-gadget/inspektor-gadget/cmd/kubectl-gadget/utils"
	gadgetv1alpha1 "github.com/inspektor-gadget/inspektor-gadget/pkg/apis/gadget/v1alpha1"
	"go.uber.org/zap"
)

const version = "v0.10.0"

func init() {
	utils.KubectlGadgetVersion, _ = semver.New(version[1:])
}

func main() {

	ctx := context.Background()

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Println("error setting up logger")
		os.Exit(1)
	}

	collect, err := collector.NewCollector(collector.CollectorOptions{Logger: logger})
	if err != nil {
		logger.Sugar().Errorw("error creating collector", "error", err.Error())
		os.Exit(1)
	}

	err = cleanTracesFromPreviousRun(ctx)
	if err != nil {
		logger.Sugar().Errorw("error cleaning up previous traces", "error", err.Error())
		os.Exit(1)
	}

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":2112", nil)

	err = collect.Collect(ctx)
	if err != nil {
		logger.Sugar().Errorw("error collecting traces", "error", err.Error())
		os.Exit(1)
	}
}

func cleanTracesFromPreviousRun(ctx context.Context) error {

	config := ctrl.GetConfigOrDie()
	cli, err := client.New(config, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		return err
	}

	opts := []client.DeleteAllOfOption{
		client.InNamespace("gadget"),
	}

	err = cli.DeleteAllOf(ctx, &gadgetv1alpha1.Trace{}, opts...)
	if err != nil {
		return err
	}

	return nil
}
