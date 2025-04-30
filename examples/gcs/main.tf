terraform {
  required_providers {
    # Use the local name defined in ~/.terraformrc
    opendaltf = {
      source = "jjkoh95/opendaltf"
    }
  }
}

provider "opendaltf" {
  service_type = "gcs"
  config = {
    "bucket" : var.bucket,
    "root" : "/test/opendal",
    "credential" : var.credential,
  }
}

resource "opendaltf_bucket" "my_fs_folder" {
  name = "some-random-dir"
}
