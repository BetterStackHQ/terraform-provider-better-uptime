provider "betteruptime" {
  api_token = var.betteruptime_api_token
}

### On-call team catalog relation

resource "betteruptime_catalog_relation" "on_call_team" {
  name        = "On-call team"
  description = "Teams with on-call responsibilities"
}

resource "betteruptime_catalog_attribute" "on_call_team_name" {
  relation_id = betteruptime_catalog_relation.on_call_team.id
  name        = "On-call team"
  primary     = true
}

resource "betteruptime_catalog_attribute" "on_call_team_lead" {
  relation_id = betteruptime_catalog_relation.on_call_team.id
  name        = "Team lead"
}

resource "betteruptime_catalog_attribute" "on_call_team_business_unit" {
  relation_id = betteruptime_catalog_relation.on_call_team.id
  name        = "Business unit"
}

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

  attribute {
    attribute_id = betteruptime_catalog_attribute.on_call_team_business_unit.id
    type         = "String"
    value        = "Customer success"
  }
}

resource "betteruptime_catalog_record" "backend_team" {
  relation_id = betteruptime_catalog_relation.on_call_team.id

  attribute {
    attribute_id = betteruptime_catalog_attribute.on_call_team_name.id
    type         = "String"
    value        = "Backend team"
  }

  attribute {
    attribute_id = betteruptime_catalog_attribute.on_call_team_lead.id
    type         = "User"
    email        = "juraj@betterstack.com"
  }

  attribute {
    attribute_id = betteruptime_catalog_attribute.on_call_team_business_unit.id
    type         = "String"
    value        = "Engineering"
  }
}

### Service catalog relation

resource "betteruptime_catalog_relation" "service" {
  name        = "Service"
  description = "Services with responsible teams"
}

resource "betteruptime_catalog_attribute" "affected_service" {
  relation_id = betteruptime_catalog_relation.service.id
  name        = "Affected service"
  primary     = true
}

# Creating a reference to On-call team by using the same name as its primary attribute
resource "betteruptime_catalog_attribute" "service_on_call_team" {
  relation_id = betteruptime_catalog_relation.service.id
  name        = betteruptime_catalog_attribute.on_call_team_name.name
}

resource "betteruptime_catalog_record" "homepage" {
  relation_id = betteruptime_catalog_relation.service.id

  attribute {
    attribute_id = betteruptime_catalog_attribute.affected_service.id
    type         = "String"
    value        = "Homepage"
  }

  attribute {
    attribute_id = betteruptime_catalog_attribute.service_on_call_team.id
    type         = "String"
    value        = "Backend team"
  }
}

resource "betteruptime_catalog_record" "api" {
  relation_id = betteruptime_catalog_relation.service.id

  attribute {
    attribute_id = betteruptime_catalog_attribute.affected_service.id
    type         = "String"
    value        = "API Services"
  }

  attribute {
    attribute_id = betteruptime_catalog_attribute.service_on_call_team.id
    type         = "String"
    value        = "Backend team"
  }
}

resource "betteruptime_catalog_record" "landing_page" {
  relation_id = betteruptime_catalog_relation.service.id

  attribute {
    attribute_id = betteruptime_catalog_attribute.affected_service.id
    type         = "String"
    value        = "Landing page"
  }

  attribute {
    attribute_id = betteruptime_catalog_attribute.service_on_call_team.id
    type         = "String"
    value        = "Demo team"
  }
}
