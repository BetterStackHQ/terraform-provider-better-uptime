---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "betteruptime_outgoing_webhook Resource - terraform-provider-better-uptime"
subcategory: ""
description: |-
  
---

# betteruptime_outgoing_webhook (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **trigger_type** (String) The type of trigger for the webhook. Only settable during creation. Available values: `incident_change`, `on_call_change`, `monitor_change`.
- **url** (String) The URL to send webhooks to.

### Optional

- **custom_webhook_template_attributes** (Block List, Max: 1) Custom webhook template configuration. (see [below for nested schema](#nestedblock--custom_webhook_template_attributes))
- **name** (String) The name of the outgoing webhook.
- **on_incident_acknowledged** (Boolean) Whether to trigger webhook when incident is acknowledged. Only when `trigger_type=incident_change`.
- **on_incident_resolved** (Boolean) Whether to trigger webhook when incident is resolved. Only when `trigger_type=incident_change`.
- **on_incident_started** (Boolean) Whether to trigger webhook when incident starts. Only when `trigger_type=incident_change`.
- **team_name** (String) Used to specify the team the resource should be created in when using global tokens.

### Read-Only

- **id** (String) The ID of the outgoing webhook.

<a id="nestedblock--custom_webhook_template_attributes"></a>
### Nested Schema for `custom_webhook_template_attributes`

Optional:

- **auth_password** (String, Sensitive)
- **auth_user** (String)
- **body_template** (String)
- **headers_template** (Block List) (see [below for nested schema](#nestedblock--custom_webhook_template_attributes--headers_template))
- **http_method** (String)

Read-Only:

- **id** (String) The ID of this resource.

<a id="nestedblock--custom_webhook_template_attributes--headers_template"></a>
### Nested Schema for `custom_webhook_template_attributes.headers_template`

Required:

- **name** (String)
- **value** (String)

