package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/pointer"
)

func main() {
	var ns string
	flag.StringVar(&ns, "ns", "default", "namespace")
	flag.Parse()
	log.Printf("start to check daprd sidecar injected failed pods for namespace [%s]...\n", ns)

	config := getK8sConfig()

	// Create an rest client not targeting specific API version
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	pods, err := clientset.CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatalln("failed to get pods:", err)
	}

	// kill pod which be injected daprd sidecar failed
	for i, pod := range pods.Items {
		daprEnabled := isPodDaprEnabled(&pod)
		if daprEnabled {
			name := pod.GetName()
			fmt.Printf("[%d] %s should have daprd sidecar\n", i, name)
			daprFound := isDardSidecarFound(&pod)
			if !daprFound {
				deleteFailedPod(i, name, clientset, ns)
			}
		}
	}

	fmt.Println("proccessed!")
}

func deleteFailedPod(i int, name string, clientset *kubernetes.Clientset, ns string) {
	fmt.Printf("-- [%d] %s injected daprd sidecar failed, will be killed\n", i, name)

	deleteOptions := metav1.DeleteOptions{
		GracePeriodSeconds: pointer.Int64(30),
		PropagationPolicy:  &[]metav1.DeletionPropagation{metav1.DeletePropagationBackground}[0],
	}
	err := clientset.CoreV1().Pods(ns).Delete(context.Background(), name, deleteOptions)
	if err == nil {
		fmt.Printf("---- [%d] %s killed\n", i, name)
	} else {
		log.Fatalf("---- [%d] %s failed to kill: %s\n", i, name, err)
	}
}

func isDardSidecarFound(pod *v1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if container.Name == "daprd" {
			return true
		}
	}
	return false
}

func isPodDaprEnabled(pod *v1.Pod) bool {
	for anno := range pod.Annotations {
		if anno == "dapr.io/enabled" {
			return true
		}
	}
	return false
}

// getK8sConfig returns a kubernetes config from InCluster or config file
func getK8sConfig() *rest.Config {
	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		log.Println("Using kubeconfig file: ", kubeconfig)

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatal(err)
		}
	}
	return config
}
