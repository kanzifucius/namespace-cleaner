package commands

import (
	"context"
	"fmt"
	gh "github.com/NovataInc/namespace-cleaner/packages/github"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type DeleteNamespacesClosedPrsCmd struct {
	CommandStruct
	Log       *zap.Logger
	ClientSet kubernetes.Interface
	Gh        *gh.GitHub
}

func NewDeleteNamespacesClosedPrsCmd(logger *zap.Logger, clientSet kubernetes.Interface, gh *gh.GitHub, dryRun bool) Command {

	cmd := &DeleteNamespacesClosedPrsCmd{
		Log:       logger,
		ClientSet: clientSet,
		Gh:        gh,
	}
	cmd.dryRun = dryRun
	return cmd
}

func (cmd *DeleteNamespacesClosedPrsCmd) Execute() error {
	namespaceClient := cmd.ClientSet.CoreV1().Namespaces()
	list, err := namespaceClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		cmd.Log.Error("Error getting namespaces", zap.Error(err))
	}
	filteredNamespaces := cmd.filterNsByAnnotation(list, []string{CanDeleteAnnotation, PrNumberAnnotation})

	for _, ns := range filteredNamespaces {
		cmd.Log.Info(fmt.Sprintf("processing ns %s for old pr deletions %v", ns.Name, ns.Annotations))
		canDelete := cmd.GetAnnotationValueIntBool(ns, CanDeleteAnnotation)
		prNumber := cmd.GetAnnotationValueInt(ns, PrNumberAnnotation)

		if prNumber > 0 && canDelete {

			merged, err := cmd.Gh.IsPrMerged(prNumber)
			if err != nil {
				return err
			}
			if merged {
				cmd.Log.Info(fmt.Sprintf("deleting ns %s as no open pr found annotaions %v dryrun : %v", ns.Name, ns.Annotations, cmd.dryRun))
				if cmd.dryRun {
					return nil
				}
				err := namespaceClient.Delete(context.TODO(), ns.Name, metav1.DeleteOptions{})
				if err != nil {
					cmd.Log.Error("error deleting namespace", zap.Error(err))
				}
				cmd.Gh.CommentOnPr("Automated NS Cleaner : Delete preview previews "+ns.Name+" as it is pr is merged", prNumber)
			}

			closed, err := cmd.Gh.IsPrClosed(prNumber)
			if err != nil {
				return err
			}

			if closed {
				cmd.Log.Info(fmt.Sprintf("deleting ns %s as no open pr found annotaions %v  dryrun : %v", ns.Name, ns.Annotations, cmd.dryRun))
				if cmd.dryRun {
					return nil
				}
				err := namespaceClient.Delete(context.TODO(), ns.Name, metav1.DeleteOptions{})
				if err != nil {
					cmd.Log.Error("error deleting namespace", zap.Error(err))
				}
				cmd.Gh.CommentOnPr("Automated NS Cleaner : Delete preview previews "+ns.Name+" as it is pr is closed", prNumber)
			}

		}
	}

	return nil
}
