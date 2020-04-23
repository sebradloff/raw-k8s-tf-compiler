resource "k8s_manifest" "default-apps_v1-Deployment-nginx-deployment" {
  content = file("${path.module}../testdata/k8s-files/one-obj.yaml")
}

