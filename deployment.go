package main
 
import (
	"bufio"
	"context"
	"flag"
	//"log"
	"fmt"
	"os"
	"path/filepath"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/homedir"
	"os/signal"
	"path"
	"syscall"
	"time"
)

var (
	namespace = "default"
)
 
func main() {
	outsideCluster := flag.Bool("outside-cluster", false, "set to true when run out of cluster. (default: false)")
	flag.Parse()

	var clientset *kubernetes.Clientset
	if *outsideCluster {
		// creates the out-cluster config
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		config, err := clientcmd.BuildConfigFromFlags("", path.Join(home, ".kube/config"))
		if err != nil {
			panic(err.Error())
		}
		// creates the clientset
		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}
	} else {
		// creates the in-cluster config
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		// creates the clientset
		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}
	}

	cm := createDeployment(clientset)
	sm := createService(clientset)

	go func() {
		for {
			read, err := clientset.
				AppsV1().
				Deployments(namespace).
				Get(
					context.Background(),
					cm.GetName(),
					metav1.GetOptions{},
				)
			if err != nil {
				panic(err.Error())
			}

			fmt.Printf("Read Pod %s/%s\n", namespace, read.GetName())
			time.Sleep(5 * time.Second)
		}
	}()

	fmt.Println("Waiting for Kill Signal...")
	var stopChan = make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-stopChan

	//fmt.Printf("Delete Pod %s/%s ", namespace, cm.GetName())
	deleteDeployment(clientset, cm)
	deleteService(clientset, sm)
}
 
func createDeployment(client kubernetes.Interface) *appsv1.Deployment {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
 
	// Create Deployment

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ntcu-nginx",
			Labels: map[string]string{
				"ntcu-k8s": "hw2",
            },
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"ntcu-k8s": "hw2",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"ntcu-k8s": "hw2",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "nginx",
							Image: "nginx:1.12",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	fmt.Println("Creating deployment...")
	result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
	return deployment
}

func createService(client kubernetes.Interface, ) *apiv1.Service {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig2", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig2", "", "absolute path to the kubeconfig file")
	}
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
 
	serviceClient := clientset.CoreV1().Services(apiv1.NamespaceDefault)
 
	// Create Deployment

	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ntcu-nginx",
			Labels: map[string]string{
				"ntcu-k8s": "hw2",
            },
		},
		Spec: apiv1.ServiceSpec{
			Selector:  map[string]string{
					"ntcu-k8s": "hw2",
			},
			Type: apiv1.ServiceTypeNodePort,
            Ports: []apiv1.ServicePort{
                {
                    Name:     "http",
                    Port:     80,
                    Protocol: apiv1.ProtocolTCP,
                },
            },
		},
	}
	fmt.Println("Creating service...")
	result, err := serviceClient.Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created service %q.\n", result.GetObjectMeta().GetName())
	return service
}

func deleteDeployment(client kubernetes.Interface, cm *appsv1.Deployment)  {
	err := client.
		AppsV1().
		Deployments(namespace).
		Delete(
			context.Background(),
			cm.GetName(),
			metav1.DeleteOptions{},
		)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Deleted Deployment %s/%s\n", cm.GetNamespace(), cm.GetName())
}

func deleteService(client kubernetes.Interface, sm *apiv1.Service)  {
	err := client.
		CoreV1().
		Services(namespace).
		Delete(
			context.TODO(),
			sm.GetName(),
			metav1.DeleteOptions{},
		)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Deleted Service %s/%s\n", sm.GetNamespace(), sm.GetName())
}
 
func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}
 
func int32Ptr(i int32) *int32 { return &i }