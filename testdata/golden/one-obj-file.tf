resource "k8s_manifest" "default_apps-v1_Deployment_nginx-deployment" {
  content = file("${path.module}../testdata/k8s-files/one-obj.yaml")
}

