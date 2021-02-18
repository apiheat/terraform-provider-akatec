data "akatec_netlist_ip" "terraform_test_1"{
	id = "12345_TERRAFORMTEST1"
}

output "terraform_test_1" {
  value       = data.akatec_netlist_ip.terraform_test_1.id
  description = "terraform_test_1 akamai netlist id"
}

resource "akatec_netlist_ip" "terraform_test_2" {

	name        = "terraform-test-2"
 	network     = "staging"
  	activate    = false

	group_id = 12345
	contract_id = "C-A27272"

 	description = "created-by-tf"
 	cidr_blocks = ["1.2.3.4/32", "9.8.7.6/32"]

 }
