provider "akatec" {
  edgerc = "/Users/partamonov/.edgerc"
  section = "terraform"
  ask = ""
}

data "akatec_lds_sources" "terraform_test_1"{}

output "test" {
  value = data.akatec_lds_sources.terraform_test_1
}
