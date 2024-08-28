package commands

import (
	"context"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
	"log"
	"testing"
	"time"
)

func setupSuite(tb testing.TB) func(tb testing.TB) {
	log.Println("setup suite")

	// Return a function to teardown the test
	return func(tb testing.TB) {
		log.Println("teardown suite")
	}
}

func TestDeleteOldNamespacesCmd_Execute(t *testing.T) {
	log, _ := zap.NewDevelopment()
	cleintset := testclient.NewSimpleClientset()

	nsMock := cleintset.CoreV1().Namespaces()
	_, err := nsMock.Create(context.TODO(),
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "TestOldNamespace",
				CreationTimestamp: metav1.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
				Annotations: map[string]string{
					"namespace-cleaner/delete": "true",
					"namespace-cleaner/hours":  "1",
				},
			},
		}, metav1.CreateOptions{})

	if err != nil {
		t.Errorf("failed to create ns %v", err)

	}

	_, err = nsMock.Create(context.TODO(),
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "TestNewNamespace",
				CreationTimestamp: metav1.Now(),
				Annotations: map[string]string{
					"namespace-cleaner/delete": "true",
					"namespace-cleaner/hours":  "1",
				},
			},
		}, metav1.CreateOptions{})

	currentTime := time.Now()
	tomorrow := currentTime.AddDate(0, 0, -6)

	_, err = nsMock.Create(context.TODO(),
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "TestTomorrowNamespace",
				CreationTimestamp: metav1.NewTime(tomorrow),
				Annotations: map[string]string{
					"namespace-cleaner/delete": "true",
					"namespace-cleaner/hours":  "168",
				},
			},
		}, metav1.CreateOptions{})

	if err != nil {
		t.Errorf("failed to create ns %v", err)

	}

	_, err = nsMock.Create(context.TODO(),
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "TestTomorrowNamespaceNotified",
				CreationTimestamp: metav1.NewTime(tomorrow),
				Annotations: map[string]string{
					"namespace-cleaner/delete":            "true",
					"namespace-cleaner/hours":             "168",
					"namespace-cleaner/notified-deletion": "true",
				},
			},
		}, metav1.CreateOptions{})

	if err != nil {
		t.Errorf("failed to create ns %v", err)

	}

	cleintset.CoreV1().Namespaces()
	namespacesCmd := NewDeleteOldNamespacesCmd(log, cleintset, nil, "", false)
	if err := namespacesCmd.Execute(); (err != nil) != false {
		t.Errorf("Execute() error = %v, wantErr %v", err, false)
	}
	_, err = nsMock.Get(context.TODO(), "TestOldNamespace", metav1.GetOptions{})
	if err == nil {
		t.Errorf("TestOldNamespace was not deleted")
	}
	_, err = nsMock.Get(context.TODO(), "TestNewNamespace", metav1.GetOptions{})
	if err != nil {
		t.Errorf("TestNewNamespace was  deleted")
	}

}

func TestDeleteOldNamespacesNotificationCmd_Execute(t *testing.T) {
	log, _ := zap.NewDevelopment()
	cleintset := testclient.NewSimpleClientset()

	nsMock := cleintset.CoreV1().Namespaces()

	currentTime := time.Now()
	tomorrow := currentTime.AddDate(0, 0, -6)

	_, err := nsMock.Create(context.TODO(),
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "TestTomorrowNamespace",
				CreationTimestamp: metav1.NewTime(tomorrow),
				Annotations: map[string]string{
					"namespace-cleaner/delete": "true",
					"namespace-cleaner/hours":  "168",
				},
			},
		}, metav1.CreateOptions{})

	if err != nil {
		t.Errorf("failed to create ns %v", err)

	}

	_, err = nsMock.Create(context.TODO(),
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "TestTomorrowNamespaceNotified",
				CreationTimestamp: metav1.NewTime(tomorrow),
				Annotations: map[string]string{
					"namespace-cleaner/delete":            "true",
					"namespace-cleaner/hours":             "168",
					"namespace-cleaner/notified-deletion": "true",
				},
			},
		}, metav1.CreateOptions{})

	if err != nil {
		t.Errorf("failed to create ns %v", err)

	}

	cleintset.CoreV1().Namespaces()
	namespacesCmd := NewDeleteOldNamespacesCmd(log, cleintset, nil, "", false)

	if err := namespacesCmd.Execute(); (err != nil) != false {
		t.Errorf("Execute() error = %v, wantErr %v", err, false)
	}

	testNotified, err := nsMock.Get(context.TODO(), "TestTomorrowNamespace", metav1.GetOptions{})
	if err == nil {
		if testNotified.Annotations["namespace-cleaner/notified-deletion"] != "true" {
			t.Errorf("TestTomorrowNamespace was not makred notified")
		}

	}

	testAlreadyNotified, err := nsMock.Get(context.TODO(), "TestTomorrowNamespaceNotified", metav1.GetOptions{})
	if err == nil {
		if testAlreadyNotified.Annotations["namespace-cleaner/notified-deletion"] != "true" {
			t.Errorf("TestTomorrowNamespaceNotified was not notified")
		}

	}

}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
