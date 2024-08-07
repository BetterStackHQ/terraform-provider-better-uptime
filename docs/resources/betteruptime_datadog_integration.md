---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "betteruptime_datadog_integration Resource - terraform-provider-better-uptime"
subcategory: ""
description: |-
  https://betterstack.com/docs/uptime/api/datadog-integrations/
---

# betteruptime_datadog_integration (Resource)

https://betterstack.com/docs/uptime/api/datadog-integrations/



<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- **alerting_rule** (String) Should we alert only on alarms, or on both alarms and warnings. Possible values: alert, alert_and_warn.
- **call** (Boolean) Do we call the on-call person?
- **email** (Boolean) Do we send an email to the on-call person?
- **name** (String) The name of the Datadog Integration.
- **paused** (Boolean) Is the Datadog integration paused.
- **policy_id** (Number) ID of the escalation policy associated with the Datadog integration.
- **push** (Boolean) Do we send a push notification to the on-call person?
- **recovery_period** (Number) How long the alert must be up to automatically mark an incident as resolved. In seconds.
- **sms** (Boolean) Do we send an SMS to the on-call person?
- **team_name** (String) Used to specify the team the resource should be created in when using global tokens.
- **team_wait** (Number) How long we wait before escalating the incident alert to the team. In seconds.

### Read-Only

- **id** (String) The ID of the Datadog Integration.
- **webhook_url** (String) The webhook URL for the Datadog integration.


