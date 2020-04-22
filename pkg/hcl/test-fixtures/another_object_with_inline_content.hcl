resource "k8s_manifest" "test-apps_v1-Deployment-another" {
  content = <<EOT
apiVersion: apps/v1
kind: Deployment
metadata:
  name: another
  namespace: test
EOT
}

