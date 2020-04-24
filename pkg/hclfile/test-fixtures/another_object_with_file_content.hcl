resource "k8s_manifest" "test_apps-v1_Deployment_another" {
  content = file("${path.module}/fake-another.yaml")
}

