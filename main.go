package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/gidoichi/secrets-store-csi-driver-provider-infisical/auth"
	"github.com/gidoichi/secrets-store-csi-driver-provider-infisical/provider"
	"github.com/gidoichi/secrets-store-csi-driver-provider-infisical/server"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	socketPath := "/etc/kubernetes/secrets-store-csi-providers/infisical.sock"
	_ = os.MkdirAll("/etc/kubernetes/secrets-store-csi-providers", 0755)
	_ = os.Remove(socketPath)
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		panic(fmt.Errorf("unable to get kubeconfig: %v", err))
	}
	kubeClient := kubernetes.NewForConfigOrDie(kubeConfig)
	auth := auth.NewAuth(kubeClient)
	infisicalClientFactory := provider.NewInfisicalClientFactory()
	provider := server.NewCSIProviderServer(socketPath, auth, infisicalClientFactory)
	defer provider.Stop()

	if err := provider.Start(); err != nil {
		panic(fmt.Errorf("unable to start server: %v", err))
	}

	log.Printf("server started at: %s\n", socketPath)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("shutting down server")
}
