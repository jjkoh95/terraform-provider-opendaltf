terraform {
  required_providers {
    # Use the local name defined in ~/.terraformrc
    opendaltf = {
      source = "jjkoh95/opendaltf"
    }
  }
}

provider "opendaltf" {
  service_type = "fs" # Use the local filesystem backend
  config = {
    "root" = "/tmp/tf-opendal-test-buckets" # Directory where buckets will be created
  }
}

resource "opendaltf_bucket" "my_fs_folder" {
  name = "my-new-blank-folder" # This will create /tmp/opendal-tf-fs-example/my-new-blank-folder/
}
