package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"gke-cost-advisor/internal/application/usecases"
	"gke-cost-advisor/internal/domain/services"
	"gke-cost-advisor/internal/infrastructure/clients"
	"gke-cost-advisor/internal/infrastructure/repositories"
	"gke-cost-advisor/internal/presentation/handlers"
	"gke-cost-advisor/pkg/config"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize infrastructure clients
	prometheusClient := clients.NewPrometheusClient(cfg.Prometheus.Endpoint, cfg.Prometheus.Token)
	kubernetesClient, err := clients.NewKubernetesClient(cfg.Kubernetes.Kubeconfig)
	if err != nil {
		log.Printf("Warning: Kubernetes client not initialized: %v", err)
	}
	billingClient := clients.NewBillingClient(cfg.GCP.ProjectID, cfg.GCP.BillingAccountID)

	// Initialize repositories
	clusterRepo := repositories.NewClusterRepository(kubernetesClient, cfg)
	containerRepo := repositories.NewContainerRepository(kubernetesClient)
	metricsRepo := repositories.NewMetricsRepository(prometheusClient, cfg)
	pricingRepo := repositories.NewPricingRepository(billingClient, cfg)

	// Initialize domain services
	efficiencyScorer := services.NewEfficiencyScorer()
	rightSizer := services.NewRightSizer()

	// Initialize use cases
	efficiencyUseCase := usecases.NewAnalyzeCostEfficiencyUseCase(
		clusterRepo,
		metricsRepo,
		efficiencyScorer,
	)

	// Initialize handlers
	clusterHandler := handlers.NewClusterHandler(clusterRepo)
	efficiencyHandler := handlers.NewEfficiencyHandler(efficiencyUseCase)
	recommendationHandler := handlers.NewRecommendationHandler(rightSizer, metricsRepo, pricingRepo, containerRepo)
	machineFamilyHandler := handlers.NewMachineFamilyHandler(pricingRepo)

	// Setup router
	mux := http.NewServeMux()

	// Register routes
	handlers.RegisterRoutes(mux, clusterHandler, efficiencyHandler, recommendationHandler, machineFamilyHandler)

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server starting on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := server.Shutdown(nil); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited")
}
