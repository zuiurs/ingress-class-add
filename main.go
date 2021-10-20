package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	ingressClassKey      = "kubernetes.io/ingress.class"
	ingressClassInternal = "internal"
)

func main() {
	ctx := context.Background()

	homedir, _ := os.UserHomeDir()
	config, err := clientcmd.BuildConfigFromFlags("", path.Join(homedir, ".kube", "config"))
	if err != nil {
		log.Fatalf("failed to get kubeconfig: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to create clientset: %v", err)
	}

	inglist, err := clientset.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("failed to get ingress list: %v", err)
	}
	for _, ing := range inglist.Items {
		fmt.Println()
		fmt.Println("--------------------------------------------------------")
		fmt.Printf("[%s (%s)] Check start\n", ing.GetName(), ing.GetNamespace())
		fmt.Println("---- Annotaton check")

		annotations := ing.GetAnnotations()
		var hasIngressClassKey bool
		for k, v := range annotations {
			fmt.Printf("      - %s: %s\n", k, v)
			if k == ingressClassKey {
				hasIngressClassKey = true
			}
		}

		if hasIngressClassKey {
			fmt.Println("---- IngressClass annotation is already set")
			fmt.Println("---- OK")
		} else {
			fmt.Println("---- IngressClass annotation is not set")

			var addIngress bool
			sc := bufio.NewScanner(os.Stdin)
		inputloop:
			for {
				fmt.Print("Add IngressClass? [y/n]: ")
				sc.Scan()
				input := sc.Text()
				switch input {
				case "y":
					addIngress = true
					break inputloop
				case "n":
					addIngress = false
					break inputloop
				default:
					continue
				}
			}
			if addIngress {
				fmt.Println("---- Adding IngressClass annotation...")
				annotations[ingressClassKey] = ingressClassInternal
				ing.SetAnnotations(annotations)

				// NOTE: comment out to execute
				//if _, err := clientset.NetworkingV1().Ingresses(ing.GetNamespace()).Update(ctx, &ing, metav1.UpdateOptions{}); err != nil {
				//	fmt.Printf("---- Failed to add IngressClass annotation: %v", err)
				//}
				fmt.Println("---- Successfuly added IngressClass annotation")
			} else {
				fmt.Println("---- Skip this Ingress resource")
				continue
			}
		}
	}
}
