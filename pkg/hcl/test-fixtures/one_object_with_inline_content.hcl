resource "k8s_manifest" "test_apps-v1_Deployment_one" {
  content = <<EOT
apiVersion: apps/v1
kind: Deployment
metadata:
  name: one
  namespace: test
EOT
}

