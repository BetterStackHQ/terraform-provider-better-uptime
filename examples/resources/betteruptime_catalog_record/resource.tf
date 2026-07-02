resource "betteruptime_catalog_record" "demo_team" {
  relation_id = betteruptime_catalog_relation.on_call_team.id

  attribute {
    attribute_id = betteruptime_catalog_attribute.on_call_team_name.id
    type         = "String"
    value        = "Demo team"
  }
  attribute {
    attribute_id = betteruptime_catalog_attribute.on_call_team_lead.id
    type         = "User"
    email        = "petr@betterstack.com"
  }
}

# API Services incidents in Production -> Demo team

resource "betteruptime_catalog_record" "api_production" {
  relation_id = betteruptime_catalog_relation.service.id

  attribute {
    attribute_id = betteruptime_catalog_attribute.affected_service.id
    type         = "String"
    value        = "API Services"
  }
  attribute {
    attribute_id = betteruptime_catalog_attribute.service_environment.id
    type         = "String"
    value        = "Production"
  }
  attribute {
    attribute_id = betteruptime_catalog_attribute.service_on_call_team.id
    type         = "String"
    value        = "Demo team"
  }
}
