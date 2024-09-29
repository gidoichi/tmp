package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gidoichi/secrets-store-csi-driver-provider-infisical/auth"
	"github.com/gidoichi/secrets-store-csi-driver-provider-infisical/provider"
	"github.com/gidoichi/secrets-store-csi-driver-provider-infisical/server"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	runtimeVersion = "0.4.0"
	versionFlag    = flag.Bool("version", false, "print version information")
)

func main() {
	flag.Parse()
	if *versionFlag {
		fmt.Println(runtimeVersion)
		os.Exit(0)
	}

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
	provider := server.NewCSIProviderServer(runtimeVersion, socketPath, auth, infisicalClientFactory)
	defer provider.Stop()

	if err := provider.Start(); err != nil {
		panic(fmt.Errorf("unable to start server: %v", err))
	}

	log.Printf("server started at: %s\n", socketPath)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	log.Println("shutting down server")
}
