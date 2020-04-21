resource "k8s_manifest" "test-1-apps_v1-Deployment-test-1" {
  content = <<EOT
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-1
  namespace: test-1
EOT

}

