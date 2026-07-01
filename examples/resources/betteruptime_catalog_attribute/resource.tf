resource "betteruptime_catalog_attribute" "on_call_team_name" {
  relation_id = betteruptime_catalog_relation.on_call_team.id
  name        = "On-call team"
  primary     = true
}

resource "betteruptime_catalog_attribute" "on_call_team_lead" {
  relation_id = betteruptime_catalog_relation.on_call_team.id
  name        = "Team lead"
  position    = 1 # Order this attribute within the relation
}

resource "betteruptime_catalog_attribute" "affected_service" {
  relation_id = betteruptime_catalog_relation.service.id
  name        = "Affected service"
  primary     = true
}

resource "betteruptime_catalog_attribute" "service_environment" {
  relation_id = betteruptime_catalog_relation.service.id
  name        = "Environment"
  primary     = true
}

# Reference the On-call team relation by reusing its primary attribute's name.
resource "betteruptime_catalog_attribute" "service_on_call_team" {
  relation_id = betteruptime_catalog_relation.service.id
  name        = betteruptime_catalog_attribute.on_call_team_name.name
}
