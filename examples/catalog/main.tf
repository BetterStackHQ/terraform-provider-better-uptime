provider "betteruptime" {
  api_token = var.betteruptime_api_token
}

resource "betteruptime_catalog_relation" "this" {
  name        = "Team"
  description = "Catalog of organisation teams"

  attributes {
    name     = "Team"
    primary  = true
    position = 0
  }

  attributes {
    name     = "Team Lead"
    primary  = false
    position = 1
  }

  attributes {
    name     = "Color"
    primary  = false
    position = 2
  }

  records = jsonencode([
    {
      Team      = [{ type = "Team", item_id = "265568" }]
      Team_Lead = [{ type = "User", email = "ivan@betterstack.com" }]
      Color     = [{ type = "String", value = "Blue" }]
    },
    {
      Team      = [{ type = "Team", name = "Alistair's Team" }]
      Team_Lead = [{ type = "User", email = "alistair@betterstack.com" }]
      Color     = [{ type = "String", value = "Red" }]
    },
    {
      Team      = [{ type = "Team", name = "Andrej's team" }]
      Team_Lead = [{ type = "User", email = "andrei@betterstack.com" }, { type = "User", email = "andrej@betterstack.com" }]
      Color     = [{ type = "String", value = "Green" }]
    }
  ])
}
