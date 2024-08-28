package commands

import (
	"context"
	"fmt"
	gh "github.com/NovataInc/namespace-cleaner/packages/github"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strconv"
	"time"
)

type DeleteNsWithCrashLoopingPods struct {
	CommandStruct
	Log       *zap.Logger
	ClientSet kubernetes.Interface
	Gh        *gh.GitHub
}

func NewDeleteNsWithCrashLoopingPods(logger *zap.Logger, clientSet kubernetes.Interface, gh *gh.GitHub, dryRun bool) Command {

	cmd := &DeleteNsWithCrashLoopingPods{
		Log:       logger,
		ClientSet: clientSet,
		Gh:        gh,
	}
	cmd.dryRun = dryRun
	return cmd
}

func (cmd DeleteNsWithCrashLoopingPods) Execute() error {
	namespaceClient := cmd.ClientSet.CoreV1().Namespaces()
	list, err := namespaceClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		cmd.Log.Error("Error getting namespaces", zap.Error(err))
	}

	filteredNamespaces := cmd.filterNsByAnnotation(list, []string{CanDeleteCrashLoopingPods, CrashLoopingPodsTolerationHours})
	cmd.Log.Info("namespaces for candidate crashLooping pod deletion", zap.Int("namespace count", len(filteredNamespaces)))
	for _, ns := range filteredNamespaces {
		//cmdConfig := commandconfig.GetConfigFromNamespaceAnnotations(ns)
		canDeleteCrashLooping := cmd.GetAnnotationValueIntBool(ns, CanDeleteCrashLoopingPods)
		canDelete := cmd.GetAnnotationValueIntBool(ns, CanDeleteAnnotation)
		podHoursToleration := cmd.GetAnnotationValueInt(ns, CrashLoopingPodsTolerationHours)
		prNumber := cmd.GetAnnotationValueInt(ns, PrNumberAnnotation)
		cmd.Log.Info(fmt.Sprintf("Checking  ns [%s] can be deleted due to crashlooping", ns.Name))
		if canDelete && canDeleteCrashLooping && podHoursToleration > 0 {

			podsClient := cmd.ClientSet.CoreV1().Pods(ns.Name)
			podList, err := podsClient.List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				cmd.Log.Error("error listing pods", zap.Error(err))
				return err
			}

			for _, pod := range podList.Items {

				for _, podstatus := range pod.Status.ContainerStatuses {
					if podstatus.State.Waiting != nil {
						if podstatus.State.Waiting.Reason == "CrashLoopBackOff" {
							cmd.Log.Info(fmt.Sprintf("namepsace [%s] has crashlooping pod :  [%v]", ns.Name, pod.Name))

							createDate := pod.CreationTimestamp
							date := time.Now()
							diff := date.Sub(createDate.Time)
							hoursDdiff := int(diff.Hours())

							if hoursDdiff > podHoursToleration && podHoursToleration > -1 {
								cmd.Log.Info(fmt.Sprintf("deleting namepsace [%s] has crashlooping pod :  [%v] from more than %d cmd.dryRun : %v ", ns.Name, pod.Name, podHoursToleration, cmd.dryRun))
								if cmd.dryRun {
									return nil
								}
								err := namespaceClient.Delete(context.TODO(), ns.Name, metav1.DeleteOptions{})
								if err != nil {
									cmd.Log.Error(fmt.Sprintf("faield to delete ns: %v", err))
								}
								if prNumber > 0 {
									cmd.Gh.CommentOnPr(
										"Automated NS Cleaner: Deleted preview "+ns.Name+" as pods crashing for more than "+strconv.Itoa(podHoursToleration)+"  hours ,pod name "+pod.Name, prNumber)
								}
							}
						}

					}
				}
			}
		}

	}
	return nil
}
