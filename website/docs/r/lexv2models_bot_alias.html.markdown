# Resource: aws_lexv2models_bot_alias

Provides an Amazon Lex V2 Bot Alias resource. For more information see
[Amazon Lex V2 Bot Alias](https://docs.aws.amazon.com/lexv2/latest/dg/aliases.html)

## Example Usage

```hcl
resource "aws_lexv2models_bot" "example" {
  name = "example"
  data_privacy {
    child_directed = false
  }
  idle_session_ttl_in_seconds = 300
}

resource "aws_lexv2models_bot_version" "example" {
  bot_id = aws_lexv2models_bot.example.id
}

resource "aws_lexv2models_bot_alias" "example" {
  bot_id      = aws_lexv2models_bot.example.id
  bot_version = aws_lexv2models_bot_version.example.bot_version
  name        = "example"
  description = "Example bot alias"
  
  tags = {
    Environment = "production"
    Project     = "example"
  }
}
```

## Argument Reference

The following arguments are supported:

* `bot_id` - (Required) The identifier of the bot to create an alias for.
* `bot_version` - (Required) The version of the bot to associate with the alias.
* `name` - (Required) The name of the bot alias. The name must be unique within the account.
* `description` - (Optional) A description of the bot alias.
* `tags` - (Optional) Key-value map of tags for the bot alias.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the bot alias in the format `bot_alias_id:bot_id`
* `arn` - The ARN of the bot alias.
* `bot_alias_id` - The unique identifier of the bot alias.
* `bot_alias_status` - The current status of the bot alias. Valid values are: `Available`, `Creating`, `Deleting`, `Failed`.

## Import

Lex V2 Bot Aliases can be imported using the `bot_alias_id:bot_id`, e.g.,

```shell
$ terraform import aws_lexv2models_bot_alias.example ABCDEF123456:GHIJKL789012
```