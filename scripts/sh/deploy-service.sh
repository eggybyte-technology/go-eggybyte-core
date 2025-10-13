#!/bin/bash
# Deployment script for EggyByte Core applications

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
SERVICE_NAME=""
ENVIRONMENT="development"
NAMESPACE="default"
DOCKER_REGISTRY="eggybyte"
DOCKER_TAG="latest"
KUBECONFIG=""
DRY_RUN=false

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Deploy EggyByte Core applications

OPTIONS:
    -s, --service NAME     Service name (required)
    -e, --env ENV          Environment (default: development)
    -n, --namespace NS      Kubernetes namespace (default: default)
    -r, --registry REG      Docker registry (default: eggybyte)
    -t, --tag TAG           Docker tag (default: latest)
    -k, --kubeconfig FILE   Kubeconfig file path
    -d, --dry-run          Dry run mode (no actual deployment)
    -h, --help             Show this help message

EXAMPLES:
    $0 -s user-service
    $0 -s payment-service -e production -n eggybyte-system
    $0 --service auth-service --env staging --tag v1.0.0 --dry-run

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -s|--service)
            SERVICE_NAME="$2"
            shift 2
            ;;
        -e|--env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -n|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        -r|--registry)
            DOCKER_REGISTRY="$2"
            shift 2
            ;;
        -t|--tag)
            DOCKER_TAG="$2"
            shift 2
            ;;
        -k|--kubeconfig)
            KUBECONFIG="$2"
            shift 2
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Validate required parameters
if [[ -z "$SERVICE_NAME" ]]; then
    print_error "Service name is required"
    show_usage
    exit 1
fi

# Set kubeconfig if provided
if [[ -n "$KUBECONFIG" ]]; then
    export KUBECONFIG="$KUBECONFIG"
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    print_error "kubectl is not installed. Please install kubectl"
    exit 1
fi

# Check if docker is installed
if ! command -v docker &> /dev/null; then
    print_error "docker is not installed. Please install docker"
    exit 1
fi

print_info "Deploying $SERVICE_NAME to $ENVIRONMENT environment"

# Docker image name
IMAGE_NAME="$DOCKER_REGISTRY/$SERVICE_NAME:$DOCKER_TAG"

# Check if image exists locally
if ! docker image inspect "$IMAGE_NAME" &> /dev/null; then
    print_warning "Docker image $IMAGE_NAME not found locally"
    print_info "Attempting to pull from registry..."
    if ! docker pull "$IMAGE_NAME"; then
        print_error "Failed to pull image $IMAGE_NAME"
        exit 1
    fi
fi

# Create namespace if it doesn't exist
if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
    print_info "Creating namespace: $NAMESPACE"
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would create namespace $NAMESPACE"
    else
        kubectl create namespace "$NAMESPACE"
    fi
fi

# Deploy using Kubernetes manifests
DEPLOYMENT_FILE="deployments/kubernetes/deployment.yaml"

if [[ ! -f "$DEPLOYMENT_FILE" ]]; then
    print_error "Deployment file not found: $DEPLOYMENT_FILE"
    exit 1
fi

# Update image in deployment file
TEMP_DEPLOYMENT=$(mktemp)
sed "s|eggybyte/service:latest|$IMAGE_NAME|g" "$DEPLOYMENT_FILE" > "$TEMP_DEPLOYMENT"
sed -i "s|eggybyte-service|$SERVICE_NAME|g" "$TEMP_DEPLOYMENT"
sed -i "s|eggybyte-system|$NAMESPACE|g" "$TEMP_DEPLOYMENT"

print_info "Deploying to Kubernetes..."

if [[ "$DRY_RUN" == "true" ]]; then
    print_info "DRY RUN: Would deploy with the following configuration:"
    echo "Service: $SERVICE_NAME"
    echo "Environment: $ENVIRONMENT"
    echo "Namespace: $NAMESPACE"
    echo "Image: $IMAGE_NAME"
    echo ""
    print_info "Deployment manifest:"
    cat "$TEMP_DEPLOYMENT"
else
    # Apply the deployment
    if kubectl apply -f "$TEMP_DEPLOYMENT"; then
        print_success "Deployment applied successfully"
    else
        print_error "Failed to apply deployment"
        rm -f "$TEMP_DEPLOYMENT"
        exit 1
    fi
    
    # Wait for deployment to be ready
    print_info "Waiting for deployment to be ready..."
    if kubectl rollout status deployment/"$SERVICE_NAME" -n "$NAMESPACE" --timeout=300s; then
        print_success "Deployment is ready"
    else
        print_error "Deployment failed to become ready"
        rm -f "$TEMP_DEPLOYMENT"
        exit 1
    fi
    
    # Show deployment status
    print_info "Deployment status:"
    kubectl get pods -l app="$SERVICE_NAME" -n "$NAMESPACE"
    
    # Show service endpoints
    print_info "Service endpoints:"
    kubectl get svc "$SERVICE_NAME" -n "$NAMESPACE"
fi

# Cleanup
rm -f "$TEMP_DEPLOYMENT"

print_success "Deployment completed successfully!"
