#!/bin/bash

# Kubernetes Deployment Script for Social API
# This script deploys the entire application to Kubernetes

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="social-api"
BACKEND_IMAGE="${BACKEND_IMAGE:-rudy128/social-api-backend:latest}"
WHATSAPP_IMAGE="${WHATSAPP_IMAGE:-rudy128/whatsapp-service:latest}"

echo -e "${BLUE}╔═══════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   Kubernetes Deployment - Social API                 ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════╝${NC}"
echo ""

# Function to print step
print_step() {
    echo -e "${GREEN}▶${NC} $1"
}

# Function to print error
print_error() {
    echo -e "${RED}✖${NC} $1"
}

# Function to print warning
print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    print_error "kubectl is not installed. Please install it first."
    exit 1
fi

print_step "Checking Kubernetes cluster connection..."
if ! kubectl cluster-info &> /dev/null; then
    print_error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
    exit 1
fi

echo -e "${GREEN}✓${NC} Connected to cluster"
echo ""

# Ask for confirmation
echo -e "${YELLOW}Configuration:${NC}"
echo "  Namespace: ${NAMESPACE}"
echo "  Backend Image: ${BACKEND_IMAGE}"
echo "  WhatsApp Image: ${WHATSAPP_IMAGE}"
echo ""
read -p "Continue with deployment? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_warning "Deployment cancelled."
    exit 0
fi

echo ""
print_step "Step 1: Creating namespace..."
kubectl apply -f k8s/namespace.yaml
echo -e "${GREEN}✓${NC} Namespace created/updated"
echo ""

print_step "Step 2: Creating RBAC resources..."
kubectl apply -f k8s/rbac.yaml
echo -e "${GREEN}✓${NC} RBAC resources created"
echo ""

print_step "Step 3: Creating ConfigMap and Secrets..."
# Update ConfigMap with image names
sed -i.bak "s|K8S_WHATSAPP_IMAGE:.*|K8S_WHATSAPP_IMAGE: \"${WHATSAPP_IMAGE}\"|" k8s/configmap.yaml
kubectl apply -f k8s/configmap.yaml
echo -e "${GREEN}✓${NC} ConfigMap and Secrets created"
echo ""

print_warning "Make sure to update secrets in k8s/configmap.yaml with secure values for production!"
echo ""

print_step "Step 4: Deploying PostgreSQL..."
kubectl apply -f k8s/postgres.yaml
echo "  Waiting for PostgreSQL to be ready..."
kubectl wait --for=condition=ready pod -l app=postgres -n ${NAMESPACE} --timeout=300s 2>/dev/null || {
    print_warning "PostgreSQL pods not ready yet, continuing..."
}
echo -e "${GREEN}✓${NC} PostgreSQL deployed"
echo ""

print_step "Step 5: Deploying Backend..."
# Update backend deployment with image name
sed -i.bak "s|image:.*social-api-backend.*|image: ${BACKEND_IMAGE}|" k8s/backend-deployment.yaml
kubectl apply -f k8s/backend-deployment.yaml
echo "  Waiting for Backend to be ready..."
kubectl wait --for=condition=ready pod -l app=backend -n ${NAMESPACE} --timeout=300s 2>/dev/null || {
    print_warning "Backend pods not ready yet, continuing..."
}
echo -e "${GREEN}✓${NC} Backend deployed"
echo ""

print_step "Step 6: Checking deployment status..."
kubectl get all -n ${NAMESPACE}
echo ""

print_step "Step 7: Checking backend logs..."
echo "  Last 20 lines from backend:"
kubectl logs -n ${NAMESPACE} -l app=backend --tail=20 || print_warning "Could not fetch logs"
echo ""

print_step "Step 8: Getting service information..."
echo ""
echo -e "${BLUE}Backend Service (ClusterIP):${NC}"
kubectl get svc backend-service -n ${NAMESPACE} -o wide
echo ""
echo -e "${BLUE}Backend LoadBalancer:${NC}"
kubectl get svc backend-loadbalancer -n ${NAMESPACE} -o wide
echo ""
echo -e "${BLUE}PostgreSQL Service:${NC}"
kubectl get svc postgres-service -n ${NAMESPACE} -o wide
echo ""

# Check if LoadBalancer has external IP
EXTERNAL_IP=$(kubectl get svc backend-loadbalancer -n ${NAMESPACE} -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
if [ -z "$EXTERNAL_IP" ]; then
    print_warning "LoadBalancer external IP is pending. Checking for hostname..."
    EXTERNAL_IP=$(kubectl get svc backend-loadbalancer -n ${NAMESPACE} -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || echo "")
fi

if [ -n "$EXTERNAL_IP" ]; then
    echo -e "${GREEN}✓${NC} External IP/Hostname: ${EXTERNAL_IP}"
    echo ""
    echo "You can access the API at: http://${EXTERNAL_IP}"
else
    print_warning "LoadBalancer external IP not assigned yet."
    echo "Use port-forward to access the service locally:"
    echo "  kubectl port-forward -n ${NAMESPACE} svc/backend-service 8080:8080"
    echo "  Then access at: http://localhost:8080"
fi

echo ""
print_step "Testing health endpoint..."
if [ -n "$EXTERNAL_IP" ]; then
    curl -s "http://${EXTERNAL_IP}/health" || print_warning "Health check failed"
else
    print_warning "Skipping health check (no external IP)"
fi

echo ""
echo -e "${GREEN}╔═══════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║   Deployment Complete!                                 ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════╝${NC}"
echo ""

echo -e "${BLUE}Useful Commands:${NC}"
echo ""
echo "View all resources:"
echo "  kubectl get all -n ${NAMESPACE}"
echo ""
echo "View backend logs:"
echo "  kubectl logs -f -n ${NAMESPACE} -l app=backend"
echo ""
echo "View PostgreSQL logs:"
echo "  kubectl logs -f -n ${NAMESPACE} -l app=postgres"
echo ""
echo "View WhatsApp pods:"
echo "  kubectl get pods -n ${NAMESPACE} -l app=whatsapp-service"
echo ""
echo "Port forward to backend:"
echo "  kubectl port-forward -n ${NAMESPACE} svc/backend-service 8080:8080"
echo ""
echo "Delete deployment:"
echo "  kubectl delete namespace ${NAMESPACE}"
echo ""
echo "Test API:"
echo "  curl http://localhost:8080/health"
echo ""

# Cleanup backup files
rm -f k8s/configmap.yaml.bak k8s/backend-deployment.yaml.bak

echo -e "${GREEN}✓${NC} Deployment script completed successfully!"
