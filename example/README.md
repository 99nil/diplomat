- kubectl apply -f examples/test-pod.yaml

- kubectl apply -f examples/hostname.yaml

- kubectl exec -it alpine-test -- sh
  
- (在容器环境内)
  
  / # curl hostname-svc:12345

  hostname-edge-84cb45ccf4-lh89n