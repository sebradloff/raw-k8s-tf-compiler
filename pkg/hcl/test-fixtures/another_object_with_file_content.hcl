resource "k8s_manifest" "test-apps_v1-Deployment-another" {
  content = file("${path.module}fake-another.yaml")
}

