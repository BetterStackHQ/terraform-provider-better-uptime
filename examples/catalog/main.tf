provider "betteruptime" {
  api_token = var.betteruptime_api_token
}

resource "betteruptime_catalog_relation" "country" {
  name = "Country"
}

resource "betteruptime_catalog_attribute" "country_code" {
  # Primary attributes can be referenced in other relations
  relation_id = betteruptime_catalog_relation.country.id
  name        = "Country code"
  primary     = true
}

resource "betteruptime_catalog_attribute" "country_name" {
  relation_id = betteruptime_catalog_relation.country.id
  name        = "Country name"
}

resource "betteruptime_catalog_record" "germany" {
  relation_id = betteruptime_catalog_relation.country.id

  attribute {
    # String values are sent in value field
    attribute_id = betteruptime_catalog_attribute.country_code.id
    type         = "String"
    value        = "DE"
  }
  attribute {
    attribute_id = betteruptime_catalog_attribute.country_name.id
    type         = "String"
    value        = "Germany"
  }
}

resource "betteruptime_catalog_record" "czechia" {
  relation_id = betteruptime_catalog_relation.country.id

  attribute {
    attribute_id = betteruptime_catalog_attribute.country_code.id
    type         = "String"
    value        = "CZ"
  }
  attribute {
    attribute_id = betteruptime_catalog_attribute.country_name.id
    type         = "String"
    value        = "Czechia"
  }
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
  # Creating a reference to Country by using the same name as its primary attribute
  relation_id = betteruptime_catalog_relation.office.id
  name        = betteruptime_catalog_attribute.country_code.name
}

resource "betteruptime_catalog_attribute" "office_contact_person" {
  relation_id = betteruptime_catalog_relation.office.id
  name        = "Office contact"
}

resource "betteruptime_catalog_attribute" "office_schedule" {
  relation_id = betteruptime_catalog_relation.office.id
  name        = "Office on-call"
}

data "betteruptime_on_call_calendar" "primary" {
}

data "betteruptime_on_call_calendar" "prague" {
  name = "Prague On-call"
}

resource "betteruptime_catalog_record" "office_prague" {
  relation_id = betteruptime_catalog_relation.office.id

  attribute {
    attribute_id = betteruptime_catalog_attribute.office_address.id
    type         = "String"
    value        = "123 Charles Street, Prague"
  }
  attribute {
    attribute_id = betteruptime_catalog_attribute.office_country.id
    type         = "String"
    value        = "CZ"
  }
  attribute {
    # Users can be referenced using email
    attribute_id = betteruptime_catalog_attribute.office_contact_person.id
    type         = "User"
    email        = "petr@betterstack.com"
  }
  attribute {
    # Non-string values can be referenced using item_id
    attribute_id = betteruptime_catalog_attribute.office_schedule.id
    type         = "Schedule"
    item_id      = data.betteruptime_on_call_calendar.prague.id
  }
  attribute {
    # Multiple values for a single attribute can be provided
    attribute_id = betteruptime_catalog_attribute.office_schedule.id
    type         = "Schedule"
    item_id      = data.betteruptime_on_call_calendar.primary.id
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
