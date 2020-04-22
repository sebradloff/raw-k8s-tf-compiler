resource "k8s_manifest" "default-apps_v1-Deployment-nginx-deployment" {
  content = <<EOT
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: nginx-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx
        ports:
        - containerPort: 80
EOT
}

resource "k8s_manifest" "default-v1-Service-nginx" {
  content = <<EOT
apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  clusterIP: ""
  ports:
  - name: http
    port: 80
    protocol: TCP
    targetPort: http
  selector:
    app: nginx
EOT
}

