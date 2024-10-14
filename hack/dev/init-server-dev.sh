#!/usr/bin/env bash

# Set strict mode
set -euo pipefail

# Function: Check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check if necessary commands are installed
for cmd in minikube kubectl helm; do
    if ! command_exists "$cmd"; then
        echo "Error: $cmd is not installed. Please install $cmd before running this script."
        exit 1
    fi
done

# Start Minikube
echo "Starting Minikube..."
if ! minikube status >/dev/null 2>&1; then
    minikube start
else
    echo "Minikube is already running."
fi

# Install Vault
echo "Installing Vault..."
if ! helm list | grep -q "vault"; then
    helm install vault hashicorp/vault --set "server.dev.enabled=true"
else
    echo "Vault is already installed."
fi

# Install External Secrets Operator
echo "Installing External Secrets Operator..."
if ! kubectl get namespace external-secrets >/dev/null 2>&1; then
    helm install external-secrets \
        external-secrets/external-secrets \
        -n external-secrets \
        --create-namespace
else
    echo "External Secrets Operator is already installed."
fi

# Install Argo CD
echo "Installing Argo CD..."
if ! kubectl get namespace argocd >/dev/null 2>&1; then
    kubectl create namespace argocd
    kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
else
    echo "Argo CD is already installed."
fi

echo "Development environment initialization complete."
