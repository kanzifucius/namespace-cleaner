package commands

import (
	"context"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
	"testing"
	"time"
)

func TestDeleteOldPodsCmd_Execute(t *testing.T) {
	log, _ := zap.NewDevelopment()
	cleintset := testclient.NewSimpleClientset()

	nsMock := cleintset.CoreV1().Namespaces()
	nsPods := cleintset.CoreV1().Pods("TestOldPods")
	_, err := nsMock.Create(context.TODO(),
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "TestOldPods",
				CreationTimestamp: metav1.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
				Annotations: map[string]string{
					"pods-cleaner/delete":      "true",
					"pods-cleaner/hours":       "1",
					"pods-cleaner/name-prefix": "test",
				},
			},
		}, metav1.CreateOptions{})

	if err != nil {
		t.Errorf("failed to create ns %v", err)

	}

	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			Name:              "test-pod",
			Namespace:         "TestOldPods",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            "nginx",
					Image:           "nginx",
					ImagePullPolicy: "Always",
				},
			},
		},
	}

	podNew := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Now(),
			Name:              "test-pod-new",
			Namespace:         "TestOldPods",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            "nginx",
					Image:           "nginx",
					ImagePullPolicy: "Always",
				},
			},
		},
	}
	podNewDiffPrefix := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Now(),
			Name:              "diffprefix-pod-new",
			Namespace:         "TestOldPods",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            "nginx",
					Image:           "nginx",
					ImagePullPolicy: "Always",
				},
			},
		},
	}

	_, err = nsPods.Create(context.TODO(), pod, metav1.CreateOptions{})
	_, err = nsPods.Create(context.TODO(), podNew, metav1.CreateOptions{})
	_, err = nsPods.Create(context.TODO(), podNewDiffPrefix, metav1.CreateOptions{})

	if err != nil {
		t.Errorf("failed to create ns %v", err)

	}

	cleintset.CoreV1().Namespaces()

	cmd := NewDeleteOldPodsCmd(log, cleintset, false)
	if err := cmd.Execute(); (err != nil) != false {
		t.Errorf("Execute() error = %v, wantErr %v", err, false)
	}

	_, err = nsPods.Get(context.TODO(), "test-pod", metav1.GetOptions{})
	if err == nil {
		t.Errorf("test-pod was not deleted")
	}

	_, err = nsPods.Get(context.TODO(), "test-pod-new", metav1.GetOptions{})
	if err != nil {
		t.Errorf("test-pod was deleted")
	}

	_, err = nsPods.Get(context.TODO(), "test-pod-new", metav1.GetOptions{})
	if err != nil {
		t.Errorf("diffprefix-pod-new was deleted")
	}

}
