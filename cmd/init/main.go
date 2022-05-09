package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	cowboyArray "github.com/mindaugasrmn/distributed-cowboys/helpers/cowboys"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	cowboyPath = flag.String("file", "cowboys.json", "Path to cowboys file")
)

func main() {
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
		panic(err)
	}

	cowboys, err := cowboyArray.CowboysArray(*cowboyPath)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	starterPod := GenStarterServicePod()
	starterPod, err = clientset.CoreV1().Pods(starterPod.Namespace).Create(ctx, starterPod, meta.CreateOptions{})
	if err != nil {
		panic(err)
	}

	logPod := GenLogServicePod()
	logPod, err = clientset.CoreV1().Pods(logPod.Namespace).Create(ctx, logPod, meta.CreateOptions{})
	if err != nil {
		panic(err)
	}

	for i, c := range cowboys {
		pod := GenPodConfig(c.Name, c.Health, c.Damage, 50001+i)
		pod, err = clientset.CoreV1().Pods(pod.Namespace).Create(ctx, pod, meta.CreateOptions{})
		if err != nil {
			panic(err)
		}
		fmt.Printf("Pod %s initialized!\n", strings.ToLower(c.Name))
	}

}

func GenPodConfig(name string, health, damage int, port int) *core.Pod {
	return &core.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name:      strings.ToLower(name),
			Namespace: "default",
			Labels: map[string]string{
				"app":  "cowboy",
				"name": name,
			},
		},
		Spec: core.PodSpec{
			RestartPolicy: core.RestartPolicyNever,
			Containers: []core.Container{
				{
					Name:            strings.ToLower(name),
					Image:           "cowboy:latest",
					ImagePullPolicy: core.PullIfNotPresent,
					Ports: []core.ContainerPort{
						{
							Name:          "http",
							HostPort:      int32(port),
							ContainerPort: int32(port),
							Protocol:      core.ProtocolTCP,
						},
					},
					Env: []core.EnvVar{
						{
							Name:  "PORT",
							Value: strconv.Itoa(port),
						},
						{
							Name:  "NAME",
							Value: name,
						},
						{
							Name:  "HEALTH",
							Value: strconv.Itoa(health),
						},
						{
							Name:  "DAMAGE",
							Value: strconv.Itoa(damage),
						},
					},
				},
			},
		},
	}
}

func GenLogServicePod() *core.Pod {
	return &core.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name:      "logd-service",
			Namespace: "default",
			Labels: map[string]string{
				"app": "logd",
			},
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:            "logd",
					Image:           "logd:latest",
					ImagePullPolicy: core.PullIfNotPresent,
					Ports:           []core.ContainerPort{},
					Env:             []core.EnvVar{},
				},
			},
		},
	}
}

func GenStarterServicePod() *core.Pod {
	return &core.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name:      "starter",
			Namespace: "default",
			Labels: map[string]string{
				"app": "starter",
			},
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:            "starter",
					Image:           "starter",
					ImagePullPolicy: core.PullIfNotPresent,
					Env:             []core.EnvVar{},
				},
			},
		},
	}
}
