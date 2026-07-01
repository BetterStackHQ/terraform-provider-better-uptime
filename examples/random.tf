# Status page subdomains are globally unique; a random suffix keeps the combined
# example runnable in parallel and re-runnable without collisions
resource "random_id" "status_page_subdomain" {
  byte_length = 8
  prefix      = "tf-status-"
}

# Unique suffix for resources whose names or e-mails must not collide across runs
# (e.g. team member invitations).
resource "random_pet" "unique" {
  length = 2
}
