package commands

import (
	"context"
	"fmt"
	gh "github.com/NovataInc/namespace-cleaner/packages/github"
	"github.com/ashwanthkumar/slack-go-webhook"
	"k8s.io/apimachinery/pkg/types"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"strconv"
	"time"
)

type DeleteOldNamespacesCmd struct {
	CommandStruct
	Log       *zap.Logger
	ClientSet kubernetes.Interface
	Gh        *gh.GitHub

	SlackWebhook string
}

func NewDeleteOldNamespacesCmd(logger *zap.Logger, clientSet kubernetes.Interface, gh *gh.GitHub, slackWebhook string, dryRun bool) Command {

	cmd := &DeleteOldNamespacesCmd{
		Log:          logger,
		ClientSet:    clientSet,
		Gh:           gh,
		SlackWebhook: slackWebhook,
	}

	cmd.dryRun = dryRun

	return cmd
}

func (cmd DeleteOldNamespacesCmd) Execute() error {
	namespaceClient := cmd.ClientSet.CoreV1().Namespaces()
	list, err := namespaceClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		cmd.Log.Error("Error getting namespaces", zap.Error(err))
	}
	filteredNamespaces := cmd.filterNsByAnnotation(list, []string{CanDeleteAnnotation, HourAnnotation})
	cmd.Log.Info("namespaces for candidate age deletion", zap.Int("namespace count", len(filteredNamespaces)))

	var deletionNamespacesTomorrow []string
	for _, ns := range filteredNamespaces {
		cmd.Log.Info(fmt.Sprintf("processing ns %s for age based deletion  %v", ns.Name, ns.Annotations), zap.String("namespace", ns.Name))
		hoursToleration := cmd.GetAnnotationValueInt(ns, HourAnnotation)
		canDelete := cmd.GetAnnotationValueIntBool(ns, CanDeleteAnnotation)
		prNumber := cmd.GetAnnotationValueInt(ns, PrNumberAnnotation)
		if canDelete && hoursToleration > 0 {
			createDate := ns.CreationTimestamp
			currentTime := time.Now()
			tomorrow := currentTime.AddDate(0, 0, 1)
			diff := currentTime.Sub(createDate.Time)
			hoursOld := int(diff.Hours())

			deletionDate := createDate.Add(time.Duration(hoursToleration) * time.Hour)

			if deletionDate.Day() == tomorrow.Day() && deletionDate.Month() == tomorrow.Month() && deletionDate.Year() == tomorrow.Year() {
				if ns.Annotations[NsNotifiedDeletion] == "true" {
					cmd.Log.Info("Namespace already notified for deletion", zap.String("namespace", ns.Name))
					continue
				}
				deletionNamespacesTomorrow = append(deletionNamespacesTomorrow, ns.Name)
				cmd.Log.Info(fmt.Sprintf("ns %s will be deleted tomorrow", ns.Name), zap.String("namespace", ns.Name))

				_, err := namespaceClient.Patch(context.TODO(), ns.Name, types.MergePatchType, []byte(fmt.Sprintf(`{"metadata":{"annotations":{"%s":"true"}}}`, NsNotifiedDeletion)), metav1.PatchOptions{})
				if err != nil {
					cmd.Log.Error("Error pathing namespaces", zap.Error(err))
					return err
				}

			}

			if hoursOld > hoursToleration && hoursToleration > -1 {
				cmd.Log.Info(fmt.Sprintf("Deleting ns  %s as is %d hours old ", ns.Name, hoursToleration), zap.String("namespace", ns.Name))
				err := namespaceClient.Delete(context.TODO(), ns.Name, metav1.DeleteOptions{})
				if err != nil {
					cmd.Log.Error(fmt.Sprintf("faield to delete ns: %v", err), zap.String("namespace", ns.Name))
				}
				if prNumber > 0 {
					cmd.Gh.CommentOnPr(
						"Automated NS Cleaner: Deleted preview "+ns.Name+" as it is more than"+strconv.Itoa(hoursToleration)+" hours old", prNumber)
				}
			}

		}
	}

	webhookUrl := cmd.SlackWebhook
	if webhookUrl == "" {
		cmd.Log.Info("Slack webhook not set, skipping slack notification")
		return nil
	}

	if len(deletionNamespacesTomorrow) > 0 {
		attachment1 := slack.Attachment{}
		for _, ns := range deletionNamespacesTomorrow {
			attachment1.AddField(slack.Field{Title: "Namespace", Value: ns})
		}
		payload := slack.Payload{
			Text:        "Namespaces to be deleted tomorrow",
			Attachments: []slack.Attachment{attachment1},
		}
		errors := slack.Send(webhookUrl, "", payload)
		if len(errors) > 0 {
			cmd.Log.Error(fmt.Sprintf("error: %s\n", errors))
		}
	}

	return nil
}
