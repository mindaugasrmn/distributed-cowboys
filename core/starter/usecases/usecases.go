package starter

import (
	"context"
	"fmt"
	"log"
	"time"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type starterUsecases struct {
}

func NewStarterUsecases() StarterUsecases {
	return &starterUsecases{}
}

type StarterUsecases interface {
	GetPods(attenders int) (*core.PodList, error)
}

func (s *starterUsecases) GetPods(attenders int) (*core.PodList, error) {
	time.Sleep(10 * time.Second)
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to initialize k8s cluster config, got error: %s", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to initialize k8s client set, got error: %s", err)
	}
	log.Println("created k8s client")
	var podList *core.PodList
	getPods := func() error {
		podList, err = clientset.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=cowboy"),
			FieldSelector: "status.phase=Running",
		})
		if err != nil {
			return err
		}
		if len(podList.Items) == 0 {
			return fmt.Errorf("failed to get cowboys from k8s api")
		}
		if podList.Items[0].Status.PodIP == "" {
			return fmt.Errorf("failed to get cowboy ip from k8s api")
		}
		if len(podList.Items[0].Spec.Containers) == 0 ||
			len(podList.Items[0].Spec.Containers[0].Ports) == 0 {
			return fmt.Errorf("failed to get cowboy container port from k8s api")
		}
		log.Println("got attenders", len(podList.Items))
		return nil
	}
	err = retry(7, 2*time.Second, getPods)
	if err != nil {
		return nil, err
	}
	// 	return nil
	// }
	// err := retry(5, time.Second, getPods)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get attenders list")
	// }
	if podList != nil && podList.Items != nil {
		log.Printf("got num %d of attenders", len(podList.Items))
		if len(podList.Items) != attenders {
			log.Println("retrying")
			err = retry(7, 2*time.Second, getPods)
		}
	}

	return podList, nil
}

func retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			log.Println("retrying after error:", err)
			time.Sleep(sleep)
			sleep *= 2
		}
		err = f()
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}
