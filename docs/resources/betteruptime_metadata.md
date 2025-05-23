---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "betteruptime_metadata Resource - terraform-provider-better-uptime"
subcategory: ""
description: |-
  https://betterstack.com/docs/uptime/api/metadata/
---

# betteruptime_metadata (Resource)

https://betterstack.com/docs/uptime/api/metadata/



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `key` (String) The key of this Metadata.
- `owner_id` (String) The ID of the owner of this Metadata.
- `owner_type` (String) The type of the owner of this Metadata. Valid values: `Monitor`, `Heartbeat`, `Incident`, `WebhookIntegration`, `EmailIntegration`, `IncomingWebhook`, `CallRouting`

### Optional

- `metadata_value` (Block List) An array of typed metadata values of this Metadata. (see [below for nested schema](#nestedblock--metadata_value))
- `team_name` (String, Deprecated) Used to specify the team the resource should be created in when using global tokens. This field is deprecated, team name doesn't have to be specified for this resource anymore.
- `value` (String, Deprecated) The value of this Metadata. This field is deprecated, use repeatable block metadata_value to define values with types instead.

### Read-Only

- `created_at` (String) The time when this metadata was created.
- `id` (String) The ID of this Metadata.
- `updated_at` (String) The time when this metadata was updated.

<a id="nestedblock--metadata_value"></a>
### Nested Schema for `metadata_value`

Optional:

- `email` (String) Email of the referenced user when type is `User`.
- `item_id` (String) ID of the referenced item when type is different than `String`.
- `name` (String) Name of the referenced item when type is different than `String`.
- `type` (String) Value types can be grouped into 2 main categories:
  - **Scalar**: `String`
  - **Reference**: `User`, `Team`, `Policy`, `Schedule`, `SlackIntegration`, `LinearIntegration`, `JiraIntegration`, `MicrosoftTeamsWebhook`, `ZapierWebhook`, `NativeWebhook`, `PagerDutyWebhook`
  
  The value of a **Scalar** type is defined using the value field.
  
  The value of a **Reference** type is defined using one of the following fields:
  - `item_id` - great choice when you know the ID of the target item.
  - `email` - your go-to choice when you're referencing users.
  - `name` - can be used to reference other items like teams, policies, etc.
  
  **The reference types require the presence of at least one of the three fields: `item_id`, `name`, `email`.**
- `value` (String) Value when type is String.


