resource "betteruptime_catalog_relation" "on_call_team" {
  name        = "On-call team"
  description = "Teams with on-call responsibilities"
}

resource "betteruptime_catalog_relation" "service" {
  name        = "Service"
  description = "Services with responsible teams"
  # Records enrich only incidents matching all primary attribute values; an empty primary value matches any value
  match_mode = "all"
}
