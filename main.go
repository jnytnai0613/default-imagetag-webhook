package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sirupsen/logrus"
	kwhhttp "github.com/slok/kubewebhook/v2/pkg/http"
	kwhlogrus "github.com/slok/kubewebhook/v2/pkg/log/logrus"
	kwhmodel "github.com/slok/kubewebhook/v2/pkg/model"
	kwhmutating "github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
)

func imageTagMutator(_ context.Context, _ *kwhmodel.AdmissionReview, obj metav1.Object) (*kwhmutating.MutatorResult, error) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return &kwhmutating.MutatorResult{}, nil
	}

	stopLen := 1
	for i, c := range pod.Spec.Containers {
		s := strings.Split(c.Image, ":")
		if len(s) == stopLen {
			pod.Spec.Containers[i].Image = fmt.Sprintf("%s%s", c.Image, ":latest")
		}
	}

	return &kwhmutating.MutatorResult{
		MutatedObject: pod,
	}, nil
}

func main() {
	logrusLogEntry := logrus.NewEntry(logrus.New())
	logrusLogEntry.Logger.SetLevel(logrus.DebugLevel)
	logger := kwhlogrus.NewLogrus(logrusLogEntry)

	// Create our mutator
	mt := kwhmutating.MutatorFunc(imageTagMutator)

	mcfg := kwhmutating.WebhookConfig{
		ID:      "imageTagMutator",
		Obj:     &corev1.Pod{},
		Mutator: mt,
		Logger:  logger,
	}
	wh, err := kwhmutating.NewWebhook(mcfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating webhook: %s", err)
		os.Exit(1)
	}

	// Get the handler for our webhook.
	whHandler, err := kwhhttp.HandlerFor(kwhhttp.HandlerConfig{Webhook: wh, Logger: logger})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating webhook handler: %s", err)
		os.Exit(1)
	}
	certFile := "/etc/webhook/certs/tls.crt"
	keyFile := "/etc/webhook/certs/tls.key"
	logger.Infof("Listening on :8080")
	err = http.ListenAndServeTLS(":8080", certFile, keyFile, whHandler)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error serving webhook: %s", err)
		os.Exit(1)
	}
}
