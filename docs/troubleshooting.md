# Troubleshooting Summary - SRE Assessment Platform

## Issues Found and Resolution Steps

### Issue 1: Frontend Service Not Accessible (503 Error)

**Root Cause:** Service selector mismatch
- Frontend service selector: `app: frontend-app`
- Frontend pod labels: `app: frontend`
- Result: Service had no endpoints

**Fix:**
```bash
kubectl patch svc frontend -n boutique -p '{"spec":{"selector":{"app":"frontend"}}}'
```

---

### Issue 2: Frontend Deployment Corrupted

**Root Cause:** Frontend deployment only contained rum-injector container, missing the actual frontend application
- Only rum-injector running on port 8080
- No frontend application container
- Result: Connection refused errors from ingress controller

**Fix:**
```bash
# Delete corrupted deployment
kubectl delete deployment frontend -n boutique

# Recreate from original manifest
kubectl apply -f boutique.yaml --namespace boutique
```

---

### Issue 3: Frontend Service Port Mismatch

**Root Cause:** Service targeting wrong port
- Service was targeting port 80
- Frontend application runs on port 8080

**Fix:**
```bash
kubectl patch svc frontend -n boutique -p '{"spec":{"ports":[{"port":80,"targetPort":8080}]}}'
```

---

### Issue 4: Kibana Returning 404

**Root Cause:** Ingress rewrite-target annotation conflicting with Kibana's basePath configuration
- Kibana configured with `server.basePath: "/kibana"`
- Ingress had `nginx.ingress.kubernetes.io/rewrite-target: /`
- Result: Paths were being rewritten incorrectly

**Fix:**
```bash
kubectl annotate ingress monitoring-ingress -n monitoring \
  nginx.ingress.kubernetes.io/rewrite-target- --overwrite
```

---

### Issue 5: Online Boutique Showing Nginx Default Page

**Root Cause:** Similar ingress rewrite issue + browser caching
- Ingress rewrite-target annotation causing routing issues
- Browser cached old nginx error page

**Fix:**
```bash
# Remove rewrite annotation
kubectl annotate ingress platform-ingress -n boutique \
  nginx.ingress.kubernetes.io/rewrite-target- --overwrite

# Client-side: Hard refresh browser (Ctrl+Shift+R)
```

---

### Issue 6: RUM Endpoint Returning 404

**Root Cause:** Path rewrite needed for RUM injector
- RUM injector expects requests at root `/`
- Ingress was sending `/rum` without rewriting

**Fix:**
```bash
kubectl patch ingress platform-ingress -n boutique --type=json \
  -p='[{"op": "add", "path": "/metadata/annotations/nginx.ingress.kubernetes.io~1rewrite-target", "value": "/$2"}]'
```

---

### Issue 7: Orphaned Deployment Cleanup

**Root Cause:** frontend-app deployment crash-looping
- Missing required environment variable: `PRODUCT_CATALOG_SERVICE_ADDR`
- Leftover from previous configuration attempts

**Fix:**
```bash
kubectl delete deployment frontend-app -n boutique
```

---

## Verification Commands

```bash
# Check all endpoints
curl -I http://52.188.176.204                    # Should return 200
curl -I http://52.188.176.204/kibana             # Should return 302
curl -I http://52.188.176.204/rum                # Should return 200
curl -I http://52.188.176.204/apm                # Returns 404 (expected)

# Verify pod status
kubectl get pods -n boutique
kubectl get pods -n monitoring

# Check service endpoints
kubectl get endpoints frontend -n boutique
kubectl get endpoints kibana-kibana -n monitoring
kubectl get endpoints apm-server-apm-server -n monitoring

# Get Kibana password
kubectl get secret -n monitoring elasticsearch-master-credentials \
  -o jsonpath='{.data.password}' | base64 -d
```

---

## Final Configuration

**Working Endpoints:**
- Online Boutique: http://52.188.176.204
- Kibana: http://52.188.176.204/kibana (Username: elastic, Password: EJTBz0UxOnhgSoA6)
- RUM Test: http://52.188.176.204/rum
- APM Server: http://52.188.176.204/apm (API endpoint - 404 at root is normal)

**Key Lessons:**
1. Always verify service selectors match pod labels
2. Remove ingress rewrite annotations when services handle their own path routing
3. APM Server 404 at root path is expected behavior - it's an API endpoint
4. Browser caching can mask fixes - always test with hard refresh or incognito mode

---

## Timeline of Fixes

1. Fixed frontend service selector mismatch
2. Recreated corrupted frontend deployment
3. Updated frontend service port configuration
4. Removed Kibana ingress rewrite annotation
5. Removed boutique ingress rewrite annotation
6. Added RUM path rewrite configuration
7. Cleaned up orphaned frontend-app deployment

All systems operational and ready for assessment.
