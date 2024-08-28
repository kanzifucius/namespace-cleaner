package commands

import (
	"context"
	"fmt"
	gh "github.com/NovataInc/namespace-cleaner/packages/github"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
	"time"
)

type DeleteOldPodsCmd struct {
	CommandStruct
	Log       *zap.Logger
	ClientSet kubernetes.Interface
	Gh        *gh.GitHub
}

func NewDeleteOldPodsCmd(logger *zap.Logger, clientSet kubernetes.Interface, dryRun bool) Command {

	cmd := &DeleteOldPodsCmd{
		Log:       logger,
		ClientSet: clientSet,
	}

	cmd.dryRun = dryRun

	return cmd
}

func (cmd DeleteOldPodsCmd) Execute() error {
	cmd.Log.Info("Deleting old pods")

	namespaceClient := cmd.ClientSet.CoreV1().Namespaces()
	list, err := namespaceClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		cmd.Log.Error("Error getting namespaces", zap.Error(err))
	}
	filteredNamespaces := cmd.filterNsByAnnotation(list, []string{CanDeletePodAnnotation, PodHoursAnnotation, PodPrefixAnnotation})
	cmd.Log.Info("namespaces for candidate pod age deletion", zap.Int("namespace count", len(filteredNamespaces)))
	for _, ns := range filteredNamespaces {
		cmd.Log.Info(fmt.Sprintf("processing ns %s for old pods deletions %v", ns.Name, ns.Annotations))
		canDelete := cmd.GetAnnotationValueIntBool(ns, CanDeletePodAnnotation)
		podHours := cmd.GetAnnotationValueInt(ns, PodHoursAnnotation)
		pdoNamePrefix := cmd.getAnnotationValueString(ns, PodPrefixAnnotation)
		if canDelete && podHours > 0 {
			cmd.Log.Info(fmt.Sprintf("Deleting old pods in namespace %s", ns.Name))

			podClient := cmd.ClientSet.CoreV1().Pods(ns.Name)
			list, err := podClient.List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				cmd.Log.Error("error listing pods", zap.Error(err))
				return err
			}

			for _, pod := range list.Items {

				if strings.HasPrefix(pod.Name, pdoNamePrefix) {
					createDate := pod.CreationTimestamp
					date := time.Now()
					diff := date.Sub(createDate.Time)
					diffHours := int(diff.Hours())

					if diffHours > podHours {
						fmt.Printf("Delting pod  %s as is %d hours old ", pod.Name, diffHours)
						if cmd.dryRun {
							return nil
						}

						err := podClient.Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
						if err != nil {
							cmd.Log.Error(fmt.Sprintf("faield to delete pod: %v", err))
						}
					}
				}

			}

		}
	}

	cmd.summary = fmt.Sprintf("Deleted old pods in %d namespaces", len(filteredNamespaces))

	return nil
}
