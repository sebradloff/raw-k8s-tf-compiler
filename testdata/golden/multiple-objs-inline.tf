resource "k8s_manifest" "default_apps-v1_Deployment_nginx-deployment-two" {
  content = <<EOT
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx-two
  name: nginx-deployment-two
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx-two
  template:
    metadata:
      labels:
        app: nginx-two
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx
        ports:
        - containerPort: 80
EOT
}

resource "k8s_manifest" "default_v1_Service_nginx-two" {
  content = <<EOT
apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx-two
  name: nginx-two
spec:
  clusterIP: ""
  ports:
  - name: http
    port: 80
    protocol: TCP
    targetPort: http
  selector:
    app: nginx-two
EOT
}

