# List all sources by type. Type is optional
data "akatec_lds_sources" "terraform"{
  type = "edns"
}

output "lds_sources" {
  value = data.akatec_lds_sources.terraform.sources
}

# List all configurations by source type.
data "akatec_lds_configurations" "terraform"{
  source_type = "edns"
}

output "lds_configurations" {
  value = data.akatec_lds_configurations.terraform.configurations
}

# Collection data resources
data "akatec_lds_log_formats" "terraform"{
  type = "cpcode-products"
}
output "lds_log_formats" {
  value = data.akatec_lds_log_formats.terraform.formats
}

data "akatec_lds_delivery_frequencies" "terraform" {}
output "lds_delivery_frequencies" {
  value = data.akatec_lds_delivery_frequencies.terraform.frequencies
}

data "akatec_lds_delivery_thresholds" "terraform" {}
output "lds_delivery_thresholds" {
  value = data.akatec_lds_delivery_thresholds.terraform.thresholds
}

data "akatec_lds_contacts" "terraform" {}
output "lds_delivery_thresholds" {
  value = data.akatec_lds_contacts.terraform.contacts
}

data "akatec_lds_netstorage_groups" "terraform" {}
output "lds_netstorage_groups" {
  value = data.akatec_lds_netstorage_groups.terraform.netstorage_groups
}

data "akatec_lds_message_sizes" "terraform" {}
output "lds_message_sizes" {
  value = data.akatec_lds_message_sizes.terraform.message_sizes
}

data "akatec_lds_encodings" "terraform" {
  log_source_type = "cpcode-products"
  delivery_type = "httpsns4" # Optional
}
output "lds_encodings" {
  value = data.akatec_lds_encodings.terraform.configurations.encodings
}

# Individual data resorces, required to create example log delivery configuration

data "akatec_lds_log_format" "terraform"{
  type = "cpcode-products"
  name = "Combined + Web App Firewall"
}
data "akatec_lds_delivery_frequency" "terraform" {
  name = "Every 3 hours"
}
data "akatec_lds_contact" "terraform" {
  name = "akamai - phone: +31200000000"
}
data "akatec_lds_netstorage_group" "terraform" {
  name = "log-delivery"
}
data "akatec_lds_message_size" "terraform" {
  name = "25 MB (approx. 150 MB uncompressed logs)"
}
data "akatec_lds_encoding" "terraform" {
  log_source_type = "cpcode-products"
  name = "GZIP"
}

# Log delivery configuration

resource "akatec_lds_configuration" "terraform" {
  start_date = "2021-02-18"
  end_date = "2021-02-23"
  status = "active"
  log_source_id = "1-123456"
  log_source_type = "cpcode-products"

  log_format_identifier = "terraform"
  log_format_id = data.akatec_lds_log_format.terraform.id

  aggregation_type = "byLogArrival"
  delivery_frequency_id = data.akatec_lds_delivery_frequency.terraform.id

  contact_details = {
    "email_addresses" = "example@akamai.com"
    "id" = data.akatec_lds_contact.terraform.id
  }

  delivery_type = "httpsns4"
  delivery_details = {
    "cp_code" = data.akatec_lds_netstorage_group.terraform.cp_code
    "directory" = "/terraform"
    "domain_prefix" = data.akatec_lds_netstorage_group.terraform.domain_prefix
  }

  message_size_id = data.akatec_lds_message_size.terraform.id
  encoding_details = {
    "id" = data.akatec_lds_encoding.terraform.id
  }
}
