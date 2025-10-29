variable "TAG" {
  default = "latest"
}

variable "REGISTRY" {
  default = ""
}

group "default" {
  targets = ["producer", "consumer"]
}

target "producer" {
  context    = "./producer"
  dockerfile = "Dockerfile"
  tags       = ["${REGISTRY}producer:${TAG}"]
  platforms  = ["linux/amd64", "linux/arm64"]
  output     = ["type=docker"]
  
  attest = [
    "type=sbom",
    "type=provenance,mode=max"
  ]
  
  cache-from = ["type=registry,ref=${REGISTRY}producer:buildcache"]
  cache-to   = ["type=registry,ref=${REGISTRY}producer:buildcache,mode=max"]
}

target "consumer" {
  context    = "./consumer"
  dockerfile = "Dockerfile"
  tags       = ["${REGISTRY}consumer:${TAG}"]
  platforms  = ["linux/amd64", "linux/arm64"]
  output     = ["type=docker"]
  
  attest = [
    "type=sbom",
    "type=provenance,mode=max"
  ]
  
  cache-from = ["type=registry,ref=${REGISTRY}consumer:buildcache"]
  cache-to   = ["type=registry,ref=${REGISTRY}consumer:buildcache,mode=max"]
}

target "producer-sbom" {
  inherits = ["producer"]
  output   = ["type=local,dest=./sbom/producer"]
  attest   = ["type=sbom,generator=docker/buildkit-syft-scanner"]
}

target "consumer-sbom" {
  inherits = ["consumer"]
  output   = ["type=local,dest=./sbom/consumer"]
  attest   = ["type=sbom,generator=docker/buildkit-syft-scanner"]
}
