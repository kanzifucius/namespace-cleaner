package main

import (
	"flag"
	"github.com/NovataInc/namespace-cleaner/packages/commands"
	gh "github.com/NovataInc/namespace-cleaner/packages/github"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
	"strconv"
)

func connectToK8s(local bool, log *zap.Logger) *kubernetes.Clientset {

	if local {
		var kubeconfig *string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()

		config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err)
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			log.Panic("Failed to create K8s clientset")
		}

		return clientset
	} else {
		config, err := rest.InClusterConfig()
		if err != nil {
			log.Panic(err.Error())
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			log.Panic("Failed to create K8s clientset")
		}

		return clientset

	}

}

func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal

}

func LookupEnvOrBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.ParseBool(val)
		if err != nil {
			panic(err)
		}
		return v
	}
	return defaultVal
}

func main() {

	var ghToken string
	var ghOwner string
	var slackWebHook string
	var ghRepo string
	var local bool

	var cleanPods bool
	var cleanPodsDryRun bool

	var cleanPrs bool
	var cleanPrsDryRun bool

	var cleanOldNamespaces bool
	var cleanOldNamespacesDryRun bool

	var cleanCrashLoops bool
	var cleanCrashLoopsDryRun bool

	log, _ := zap.NewProduction()

	flag.BoolVar(&local, "local", LookupEnvOrBool("LOCAL", false), "Run with local kube config")

	flag.BoolVar(&cleanPods, "CLEAN_PODS", LookupEnvOrBool("CLEAN_PODS", false), "clean pods")
	flag.BoolVar(&cleanPodsDryRun, "CLEAN_PODS_DRYRUN", LookupEnvOrBool("CLEAN_PODS_DRYRUN", false), "clean pods")

	flag.BoolVar(&cleanPrs, "CLEAN_PRS", LookupEnvOrBool("CLEAN_PRS", false), "clean PRs")
	flag.BoolVar(&cleanPrsDryRun, "CLEAN_PRS_DRYRUN", LookupEnvOrBool("CLEAN_PRS_DRYRUN", false), "clean PRs")

	flag.BoolVar(&cleanOldNamespaces, "CLEAN_OLD_NS", LookupEnvOrBool("CLEAN_OLD_NS", false), "clean old namespaces")
	flag.BoolVar(&cleanOldNamespacesDryRun, "CLEAN_OLD_NS_DRYRUN", LookupEnvOrBool("CLEAN_OLD_NS_DRYRUN", false), "clean old namespaces")

	flag.BoolVar(&cleanCrashLoops, "CLEAN_CRASHLOOPS_NS", LookupEnvOrBool("CLEAN_CRASHLOOPS_NS", false), "clean ns with crashlooping pods")
	flag.BoolVar(&cleanCrashLoopsDryRun, "CLEAN_CRASHLOOPS_NS_DRYRUN", LookupEnvOrBool("CLEAN_CRASHLOOPS_NS_DRYRUN", false), "clean ns with crashlooping pods")

	flag.StringVar(&ghToken, "GH_TOKEN", LookupEnvOrString("GH_TOKEN", ""), "gh pat")
	flag.StringVar(&ghOwner, "GH_OWNER", LookupEnvOrString("GH_OWNER", ""), "gh owner")
	flag.StringVar(&ghRepo, "GH_REPO", LookupEnvOrString("GH_REPO", ""), "gh repo")
	flag.StringVar(&slackWebHook, "SLACK_WEBHOOK", LookupEnvOrString("SLACK_WEBHOOK", ""), "slack webhook ")
	flag.Parse()

	clientSet := connectToK8s(local, log)
	github := gh.NewGitHub(ghOwner, ghRepo, ghToken, log)

	cmdRunner := commands.NewCommandRunner(log)
	if cleanPrs {
		cmdRunner.AddCommand(commands.NewDeleteNamespacesClosedPrsCmd(log, clientSet, github, cleanPrsDryRun))
	}
	if cleanOldNamespaces {
		cmdRunner.AddCommand(commands.NewDeleteOldNamespacesCmd(log, clientSet, github, slackWebHook, cleanOldNamespacesDryRun))
	}
	if cleanPods {
		cmdRunner.AddCommand(commands.NewDeleteOldPodsCmd(log, clientSet, cleanPodsDryRun))
	}
	if cleanCrashLoops {
		cmdRunner.AddCommand(commands.NewDeleteNsWithCrashLoopingPods(log, clientSet, github, cleanCrashLoopsDryRun))
	}

	cmdRunner.Run()

}
