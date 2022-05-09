package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to initialize k8s cluster config, got error: %s", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to initialize k8s client set, got error: %s", err)
	}
	log.Println("created k8s client")
	podList, err := clientset.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=cowboy"),
	})
	if err != nil {
		log.Println(err)
		return
	}
	if podList != nil {
		log.Println(len(podList.Items))
		for _, pod := range podList.Items {

			value := fmt.Sprintf("%s:%d", pod.Status.PodIP, pod.Spec.Containers[0].Ports[0].ContainerPort)
			fmt.Println(value)

		}
	}
	if len(podList.Items) == 0 {
		log.Println("not found")
		return
	}
	// if podList.Items[0].Status.PodIP == "" {
	// 	return fmt.Errorf("failed to get pods ip from k8s api")
	// }
	// if len(podList.Items[0].Spec.Containers) == 0 ||
	// 	len(podList.Items[0].Spec.Containers[0].Ports) == 0 {
	// 	return fmt.Errorf("failed to get cowboy container port from k8s api")
	// }
	return

}

func structToJSONString(data interface{}) string {
	str, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(str)
}
