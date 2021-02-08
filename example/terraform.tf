provider "akatec" {
  edgerc = "/Users/rafpe/.edgerc"
  section = "test"
  ask = ""
}

data "akatec_netlist_ip" "terraform_test_1"{
	id = "12345_TERRAFORMTEST1"
}

output "terraform_test_1" {
  value       = data.akatec_netlist_ip.terraform_test_1.id
  description = "terraform_test_1 akamai netlist id"
}

resource "akatec_netlist_ip" "terraform_test_2" {
    
	name        = "terraform-test-2"
	acg 		= "xxxxxxx"
 	network     = "staging"
  	activate    = false

 	description = "created-by-tf"
 	cidr_blocks = ["1.2.3.4/32", "9.8.7.6/32"]

 }