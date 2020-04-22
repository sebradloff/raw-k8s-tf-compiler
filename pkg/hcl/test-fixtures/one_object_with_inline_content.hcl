resource "k8s_manifest" "test-apps_v1-Deployment-one" {
  content = <<EOT
apiVersion: apps/v1
kind: Deployment
metadata:
  name: one
  namespace: test
EOT
}

