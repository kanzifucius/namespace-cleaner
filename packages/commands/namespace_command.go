package commands

import (
	v1 "k8s.io/api/core/v1"
	"strconv"
)

type Command interface {
	Execute() error
	GetSummary() string
}

type CommandStruct struct {
	summary string
	dryRun  bool
}

func (cmd CommandStruct) filterNsByAnnotation(nsList *v1.NamespaceList, annotation []string) []v1.Namespace {

	var filteredNs []v1.Namespace
	for _, ns := range nsList.Items {

		if ns.Annotations != nil {
			found := true
			for _, ann := range annotation {
				if _, ok := ns.Annotations[ann]; !ok {
					found = false
				}
			}

			if found {
				filteredNs = append(filteredNs, ns)
			}
		}

	}
	return filteredNs
}

func (cmd CommandStruct) GetSummary() string {
	return cmd.summary
}

func (cmd CommandStruct) GetAnnotationValueIntBool(ns v1.Namespace, annotation string) bool {
	if ns.Annotations != nil {
		if value, ok := ns.Annotations[annotation]; ok {
			if value == "true" {
				return true
			}
		}
	}
	return false
}

func (cmd CommandStruct) GetAnnotationValueInt(ns v1.Namespace, annotation string) int {
	if ns.Annotations != nil {
		if value, ok := ns.Annotations[annotation]; ok {
			intVar, err := strconv.Atoi(value)
			if err == nil {
				return intVar
			}
		}
	}
	return -1
}

func (cmd CommandStruct) getAnnotationValueString(ns v1.Namespace, annotation string) string {
	if ns.Annotations != nil {
		if value, ok := ns.Annotations[annotation]; ok {
			return value
		}
	}
	return ""
}
