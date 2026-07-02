# Status page subdomains are globally unique; a random suffix keeps the combined
# example runnable in parallel and re-runnable without collisions
resource "random_id" "status_page_subdomain" {
  byte_length = 8
}

# Unique suffix for resources whose names must not collide across runs
resource "random_pet" "unique" {
  length = 2
}
