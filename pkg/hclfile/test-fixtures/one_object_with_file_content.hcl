resource "k8s_manifest" "test_apps-v1_Deployment_one" {
  content = file("${path.module}fake-one.yaml")
}

