provider "betteruptime" {
  api_token = var.betteruptime_api_token
}

resource "betteruptime_catalog_relation" "office" {
  name        = "Office"
  description = "A physical office building representing ACME Group"
}

resource "betteruptime_catalog_attribute" "office_address" {
  relation_id = betteruptime_catalog_relation.office.id
  name        = "Office address"
  primary     = true
}

resource "betteruptime_catalog_attribute" "office_country" {
  relation_id = betteruptime_catalog_relation.office.id
  name        = "Office country code"
  primary     = false
}

resource "betteruptime_catalog_attribute" "office_contact_person" {
  relation_id = betteruptime_catalog_relation.office.id
  name        = "Office contact"
  primary     = false
}

resource "betteruptime_catalog_attribute" "office_schedule" {
  relation_id = betteruptime_catalog_relation.office.id
  name        = "Office on-call"
  primary     = false
}

data "betteruptime_on_call_calendar" "prague" {
  name = "Prague On-call"
}

resource "betteruptime_catalog_record" "office_prague" {
  relation_id = betteruptime_catalog_relation.office.id

  attribute {
    attribute_id = betteruptime_catalog_attribute.office_country.id
    type         = "String"
    value        = "123 Charles Street, Prague"
  }
  attribute {
    attribute_id = betteruptime_catalog_attribute.office_country.id
    type         = "String"
    value        = "CZ"
  }
  attribute {
    attribute_id = betteruptime_catalog_attribute.office_contact_person.id
    type         = "User"
    email        = "petr@betterstack.com"
  }
  attribute {
    attribute_id = betteruptime_catalog_attribute.office_schedule.id
    type         = "Schedule"
    item_id      = data.betteruptime_on_call_calendar.prague.id
  }
}

data "betteruptime_on_call_calendar" "berlin" {
  name = "Berlin On-call"
}

resource "betteruptime_catalog_record" "office_berlin" {
  relation_id = betteruptime_catalog_relation.office.id

  attribute {
    attribute_id = betteruptime_catalog_attribute.office_address.id
    type         = "String"
    value        = "45 Brandenburg Gate, Berlin"
  }
  attribute {
    attribute_id = betteruptime_catalog_attribute.office_country.id
    type         = "String"
    value        = "DE"
  }
  attribute {
    attribute_id = betteruptime_catalog_attribute.office_contact_person.id
    type         = "User"
    email        = "juraj@betterstack.com"
  }
  attribute {
    attribute_id = betteruptime_catalog_attribute.office_schedule.id
    type         = "Schedule"
    item_id      = data.betteruptime_on_call_calendar.berlin.id
  }
}
