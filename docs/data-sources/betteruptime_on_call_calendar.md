---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "betteruptime_on_call_calendar Data Source - terraform-provider-better-uptime"
subcategory: ""
description: |-
  On-call calendar lookup.
---

# betteruptime_on_call_calendar (Data Source)

On-call calendar lookup.



<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- **name** (String) Name of the on-call calendar.

### Read-Only

- **default_calendar** (Boolean) Whether the on-call calendar is the default on-call calendar.
- **id** (String) The ID of the on-call calendar.
- **on_call_users** (List of Object) Array of on-call persons. (see [below for nested schema](#nestedatt--on_call_users))

<a id="nestedatt--on_call_users"></a>
### Nested Schema for `on_call_users`

Read-Only:

- **email** (String)
- **first_name** (String)
- **id** (String)
- **last_name** (String)
- **phone_numbers** (List of String)


