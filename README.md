# SRE Assessment
# MTN / ITStack SRE Practical Assessment

## Overview
Full-stack observability implementation on Azure Kubernetes Service (AKS) using the Elastic Stack, OpenTelemetry, and the Google Online Boutique microservices demo.

## Live Endpoints
| Service | URL |
|---------|-----|
| Online Boutique | http://52.188.176.204 |
| Kibana | http://52.188.176.204/kibana |
| APM Server | http://52.188.176.204/apm |

## Prerequisites
- Azure CLI (`az`) — logged in with active subscription
- `kubectl` — configured for AKS cluster
- `helm` v3.x
- `git`

## Cluster Setup
```bash
az aks get-credentials --resource-group sre-assessment --name sre-cluster
kubectl get nodes  # verify 2 nodes running
```

## Deploy Order
> ⚠️ Deploy in this exact order — collectors crash-loop if APM Server doesn't exist first.

### 1. Elastic Stack
```bash
helm repo add elastic https://helm.elastic.co && helm repo update

# Elasticsearch
helm install elasticsearch elastic/elasticsearch --namespace monitoring \
  --set replicas=1 --set minimumMasterNodes=1 \
  --set resources.requests.memory=4Gi --set resources.limits.memory=6Gi

# Get ES password
ES_PASSWORD=$(kubectl get secret -n monitoring elasticsearch-master-credentials \
  -o jsonpath='{.data.password}' | base64 -d)

# Create Kibana service account token
kubectl exec -n monitoring elasticsearch-master-0 -- \
  elasticsearch-service-tokens create elastic/kibana kibana-token

# Kibana
helm install kibana elastic/kibana --namespace monitoring --no-hooks \
  -f infrastructure/kibana/values-kibana.yaml

# APM Server
helm install apm-server elastic/apm-server --namespace monitoring \
  -f infrastructure/apm-server/values-apm-server.yaml
```

### 2. OTel Collectors
```bash
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts

helm install otel-gateway open-telemetry/opentelemetry-collector \
  --namespace monitoring --set nameOverride=gateway \
  -f otel-collector/values-gateway.yaml

helm install otel-agent open-telemetry/opentelemetry-collector \
  --namespace monitoring -f otel-collector/values-agent.yaml

# Enable hostNetwork on agent (required for node IP routing)
kubectl patch daemonset otel-agent-opentelemetry-collector-agent -n monitoring \
  --patch '{"spec":{"template":{"spec":{"hostNetwork":true,"dnsPolicy":"ClusterFirstWithHostNet"}}}}'
```

### 3. Online Boutique
```bash
kubectl create namespace boutique
kubectl apply -f boutique.yaml
```

### 4. Instrumentation
```bash
# Frontend (Go)
kubectl patch deployment frontend -n boutique \
  --patch-file instrumentation/frontend/deployment-patch.yaml

# Cartservice (C#)
kubectl patch deployment cartservice -n boutique \
  --patch-file instrumentation/cartservice/deployment-patch.yaml

# Recommendationservice (Python) — OTel built-in, just enable it
kubectl patch deployment recommendationservice -n boutique \
  --patch-file instrumentation/recommendationservice/deployment-patch.yaml
```

### 5. Infrastructure Monitoring
```bash
# VM metrics
helm install metricbeat elastic/metricbeat --namespace monitoring \
  -f infrastructure/metricbeat/values-metricbeat.yaml

# NGINX metrics
kubectl apply -f infrastructure/nginx-integration/metricbeat-nginx.yaml
kubectl apply -f infrastructure/nginx-integration/filebeat-nginx.yaml

# Redis metrics
kubectl apply -f infrastructure/redis-integration/metricbeat-redis.yaml

# K8s audit logs + network policy
kubectl apply -f infrastructure/elastic-agent-policies/filebeat-k8s-audit.yaml
kubectl apply -f infrastructure/elastic-agent-policies/elastic-agent-daemonset.yaml
```

## Verify Everything Running
```bash
kubectl get pods -n monitoring
kubectl get pods -n boutique
```

Expected: all pods `1/1 Running`

## Kibana Access
- URL: `http://52.188.176.204/kibana`
- Username: `elastic`
- Password: retrieve with:
```bash
kubectl get secret -n monitoring elasticsearch-master-credentials \
  -o jsonpath='{.data.password}' | base64 -d
```

## Architecture Decisions
See [docs/DECISIONS.md](docs/DECISIONS.md) for all Architecture Decision Records (ADR-001 to ADR-007).

## Known Issues
- Traces pipeline: Go and C# pre-built images lack compiled-in OTel SDK — traces not flowing to Kibana APM. Fix: use OTel Operator for automatic SDK injection.
- RUM: Agent deployed but ingress path routing needs fix for browser access.