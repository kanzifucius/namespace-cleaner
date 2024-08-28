# Namspace cleaner


# Description
  This is a small utility that deletes or cleans namespaces from a given set of annoations on the namespace

## Annotaions
 
| Annoation                   | Value | Description                                                         |
|-----------------------------|-------|---------------------------------------------------------------------|
| namespace-cleaner/delete    | true  | Opts into namespaces cleaner                                        |
| namespace-cleaner/hours     | int   | Number of hours namespaces from creation will be deleted            |
| namespace-cleaner/pr-number | int   | GH pr number , should the pr be closed the namesapce will be deleted |
| namespace-cleaner/delete-if-has-crashlooping-pods | true   | delete the namespace if crash looping pods exist                    |
| pods-cleaner/hours          | int   | Number of max hours a pod can live in the namespace                 |
| pods-cleaner/delete         | true  | opt into pod clean ups on namespace                                 |
| pods-cleaner/name-prefix    | int   | pre fix for pod names to conisder for deletion                      |


## Old Namespaces
    Old namespace will be delete with the annotation set.
 
```
 apiVersion: v1                                                                                                                                                                                                                                │
 kind: Namespace                                                                                                                                                                                                                               │
 metadata:                                                                                                                                                                                                                                     │
   annotations:                                                                                                                                                                                                                                │
     namespace-cleaner/delete: "true"                                                                                                                                                                                                          │
     namespace-cleaner/hours: "336"                                                                                                                                                                                                            │
     namespace-cleaner/pr-number: "2008"                                                                                                                                                                                                       │
   labels:                                                                                                                                                                                                                                     │
     kubernetes.io/metadata.name: preview-2008                                                                                                                                                                                                 │
   name: preview-2008                                                                                                                                                                                                                          │
```

## Old Pods
    Old pod will be cleaned with the annotation set.

```
 apiVersion: v1                                                                                                                                                                                                                                │
 kind: Namespace                                                                                                                                                                                                                               │
 metadata:                                                                                                                                                                                                                                     │
     pods-cleaner/hours: “4”                                                                                                                                                                                                                   │
     pods-cleaner/delete: “true”                                                                                                                                                                                                           │
     pods-cleaner/name-prefix: “harness-”                                                                                                                                                                                                  │
   name: harness-delegate-ng                                                                                                                                                                                                                   │

```
