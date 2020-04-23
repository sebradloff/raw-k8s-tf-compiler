resource "k8s_manifest" "test_apps-v1_Deployment_another" {
  content = <<EOT
apiVersion: apps/v1
kind: Deployment
metadata:
  name: another
  namespace: test
EOT
}

