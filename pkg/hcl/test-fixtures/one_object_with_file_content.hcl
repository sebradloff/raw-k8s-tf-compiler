resource "k8s_manifest" "test-apps_v1-Deployment-one" {
  content = file("${path.module}fake-one.yaml")
}

