# Architecture Decision Records

## ADR-001: AKS over GKE
**Decision:** Azure Kubernetes Service (AKS)
**Reason:** Azure credits available; no GCP credits. AKS provides equivalent K8s with
Azure Network Policy (required for Section 3.3) and auto-provisioned Load Balancers.
**Trade-off:** Azure-specific tooling, but all Elastic Stack integrations are cloud-agnostic.

## ADR-002: Tail-based Sampling at OTel Gateway
**Decision:** 10% probabilistic + 100% errors + 100% latency >2s at Gateway level.
**Reason:** Head-based sampling at the agent drops spans before context is known. Gateway
sees the complete trace before deciding — ensuring every error and SLA violation is kept.
Metrics pipeline removed from gateway due to DeadlineExceeded timeouts on APM Server 8.x.
**Trade-off:** Gateway buffers traces in memory. Mitigation: memory_limiter at 80%.

## ADR-003: Kibana Exposed via NGINX Ingress (not LoadBalancer)
**Decision:** Kibana served at http://NGINX_IP/kibana via existing NGINX Ingress.
**Reason:** Reuses existing LoadBalancer IP — no extra Azure cost. Kibana service stays
ClusterIP (not directly internet-exposed).
**Production:** Add OAuth2 proxy + Azure AD + internal LoadBalancer + VPN.

## ADR-004: Metricbeat over Elastic Agent for VM Monitoring
**Decision:** Helm-installed Metricbeat DaemonSet for system and Redis metrics.
**Reason:** Elastic Agent standalone image has entrypoint issues in AKS — elastic-agent binary
not found. Metricbeat is stable and achieves identical metric coverage.
**Production:** Elastic Agent with Fleet Server preferred for centralized policy management.

## ADR-005: Service Account Token for Kibana Authentication
**Decision:** elasticsearch-service-tokens (not elastic superuser credentials).
**Reason:** Elasticsearch 8.x forbids using the elastic superuser for Kibana.
Service account tokens are scoped to Kibana only.

## ADR-006: OTel Instrumentation via Env Vars (no image rebuild)
**Decision:** Instrument pre-built Google microservice images without rebuilding.
**Reason:** Online Boutique runs pre-built images from Google Container Registry.
Rebuilding images would require forking the repo and setting up CI/CD — out of scope.
**Trade-off:** Go and C# services lack compiled-in OTel SDK — traces not flowing.
**Production:** Use OTel Operator for automatic SDK injection without image changes.

## ADR-007: Single-node Elasticsearch
**Decision:** replicas=1, minimumMasterNodes=1 for assessment.
**Reason:** 24h assessment — durability not a concern. Multi-node ES would exhaust
the 2-node AKS cluster capacity needed for the boutique application workload.
**Production:** Minimum 3 master-eligible + 2 dedicated data nodes with ILM policies.
