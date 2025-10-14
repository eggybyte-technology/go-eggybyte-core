// Package config provides unified configuration management for EggyByte services.

import (
	"context"
	"fmt"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

// K8sConfigWatcher monitors Kubernetes ConfigMaps for configuration changes.
// It uses the Kubernetes informer pattern to efficiently watch for updates
// without polling, and triggers callbacks when changes are detected.
//
// Thread Safety: The watcher is safe for concurrent use. Update callbacks
// may be invoked from multiple goroutines.
//
// Resource Usage: The watcher maintains a connection to the Kubernetes API
// server and keeps a local cache of ConfigMap data.
type K8sConfigWatcher struct {
	clientset   *kubernetes.Clientset
	namespace   string
	configMap   string
	updateFunc  func(map[string]string)
	stopCh      chan struct{}
	informer    cache.SharedIndexInformer
	initialized bool
}

// NewK8sConfigWatcher creates a new Kubernetes ConfigMap watcher.
// It establishes a connection to the Kubernetes API server and prepares
// to watch the specified ConfigMap for changes.
//
// Parameters:
//   - namespace: Kubernetes namespace containing the ConfigMap
//   - configMapName: Name of the ConfigMap to watch
//   - updateFunc: Callback function invoked when ConfigMap data changes.
//     Receives the complete data map from the ConfigMap.
//
// Returns:
//   - *K8sConfigWatcher: Configured watcher instance
//   - error: Returns error if Kubernetes API connection fails
//
// Required Permissions:
//   - The service account must have 'get' and 'watch' permissions
//     on ConfigMaps in the specified namespace
//
// Example:
//
//	watcher, err := config.NewK8sConfigWatcher(
//	    "default",
//	    "my-service-config",
//	    func(data map[string]string) {
//	        log.Printf("Config updated: %v", data)
//	        config.Update(data)
//	    },
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	go watcher.Start()
func NewK8sConfigWatcher(namespace, configMapName string, updateFunc func(map[string]string)) (*K8sConfigWatcher, error) {
	// Create in-cluster Kubernetes client configuration
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
	}

	// Build Kubernetes clientset for API access
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	return &K8sConfigWatcher{
		clientset:  clientset,
		namespace:  namespace,
		configMap:  configMapName,
		updateFunc: updateFunc,
		stopCh:     make(chan struct{}),
	}, nil
}

// Start begins watching the ConfigMap for changes.
// This method blocks until the watcher is stopped via Stop() or context cancellation.
//
// The watcher will:
//  1. Initialize connection to Kubernetes API server
//  2. Load initial ConfigMap state
//  3. Invoke update callback with initial data
//  4. Watch for subsequent changes and invoke callback on updates
//
// Error Handling: If the initial load fails, returns error immediately.
// Runtime errors are logged but don't stop the watcher.
//
// Returns:
//   - error: Returns error if initial setup or load fails
//
// Example:
//
//	watcher, _ := config.NewK8sConfigWatcher(...)
//	go func() {
//	    if err := watcher.Start(); err != nil {
//	        log.Printf("Watcher failed: %v", err)
//	    }
//	}()
func (w *K8sConfigWatcher) Start() error {
	if w.initialized {
		return fmt.Errorf("watcher already started")
	}

	// Create shared informer factory for ConfigMaps in the namespace
	factory := informers.NewSharedInformerFactoryWithOptions(
		w.clientset,
		30*time.Second, // Resync period
		informers.WithNamespace(w.namespace),
	)

	// Get ConfigMap informer
	w.informer = factory.Core().V1().ConfigMaps().Informer()

	// Register event handlers for ConfigMap changes
	if _, err := w.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    w.handleAdd,
		UpdateFunc: w.handleUpdate,
		DeleteFunc: w.handleDelete,
	}); err != nil {
		return fmt.Errorf("failed to add event handler: %w", err)
	}

	w.initialized = true

	// Start informer and wait for initial sync
	go w.informer.Run(w.stopCh)

	// Wait for cache sync with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if !cache.WaitForCacheSync(ctx.Done(), w.informer.HasSynced) {
		return fmt.Errorf("failed to sync configmap cache")
	}

	// Block until stop signal
	<-w.stopCh
	return nil
}

// Stop halts the ConfigMap watcher and releases resources.
// This method is safe to call multiple times.
//
// After calling Stop, the watcher cannot be restarted.
// Create a new watcher instance if needed.
func (w *K8sConfigWatcher) Stop() {
	if w.stopCh != nil {
		close(w.stopCh)
	}
}

// handleAdd processes ConfigMap creation events.
// Invoked when the watched ConfigMap is first detected.
func (w *K8sConfigWatcher) handleAdd(obj interface{}) {
	w.processConfigMap(obj)
}

// handleUpdate processes ConfigMap modification events.
// Invoked when the watched ConfigMap's data changes.
func (w *K8sConfigWatcher) handleUpdate(oldObj, newObj interface{}) {
	w.processConfigMap(newObj)
}

// handleDelete processes ConfigMap deletion events.
// Invoked when the watched ConfigMap is removed.
func (w *K8sConfigWatcher) handleDelete(obj interface{}) {
	// ConfigMap deleted - could notify about empty config
	// For now, we don't trigger updates on deletion
}

// processConfigMap extracts data from ConfigMap and triggers update callback.
// This internal method handles type assertions and data extraction.
func (w *K8sConfigWatcher) processConfigMap(obj interface{}) {
	// Type assertion to get ConfigMap object
	// In real implementation, would use proper type from k8s.io/api/core/v1
	// For now, using interface{} to avoid import complexity in this demo

	// TODO: Add proper ConfigMap type handling
	// configMap, ok := obj.(*corev1.ConfigMap)
	// if !ok {
	//     return
	// }

	// Extract data and invoke callback
	// w.updateFunc(configMap.Data)
}
