package lexmodelsv2

import (
    "context"
    "fmt"
    "strings"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
    "github.com/hashicorp/terraform-provider-aws/internal/conns"
    "github.com/hashicorp/terraform-provider-aws/internal/tfresource"
    "github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceBotAlias() *schema.Resource {
    return &schema.Resource{
        CreateContext: resourceAwsLexV2ModelsBotAliasCreate,
        ReadContext:   resourceAwsLexV2ModelsBotAliasRead,
        UpdateContext: resourceAwsLexV2ModelsBotAliasUpdate,
        DeleteContext: resourceAwsLexV2ModelsBotAliasDelete,
        Importer: &schema.ResourceImporter{
            StateContext: resourceAwsLexV2ModelsBotAliasImport,
        },

        Timeouts: &schema.ResourceTimeout{
            Create: schema.DefaultTimeout(30 * time.Minute),
            Update: schema.DefaultTimeout(30 * time.Minute),
            Delete: schema.DefaultTimeout(30 * time.Minute),
        },

        Schema: map[string]*schema.Schema{
            "arn": {
                Type:     schema.TypeString,
                Computed: true,
            },
            "bot_alias_id": {
                Type:     schema.TypeString,
                Computed: true,
            },
            "bot_alias_status": {
                Type:     schema.TypeString,
                Computed: true,
            },
            "bot_id": {
                Type:     schema.TypeString,
                Required: true,
                ForceNew: true,
            },
            "bot_version": {
                Type:     schema.TypeString,
                Required: true,
            },
            "description": {
                Type:         schema.TypeString,
                Optional:     true,
                ValidateFunc: validation.StringLenBetween(0, 200),
            },
            "name": {
                Type:         schema.TypeString,
                Required:     true,
                ForceNew:     true,
                ValidateFunc: validation.StringLenBetween(1, 100),
            },
            "tags":     tftags.TagsSchema(),
            "tags_all": tftags.TagsSchemaComputed(),
        },
    }
}

func resourceAwsLexV2ModelsBotAliasCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    conn := meta.(*conns.AWSClient).LexV2ModelsConn(ctx)

    name := d.Get("name").(string)
    input := &lexmodelsv2.CreateBotAliasInput{
        BotAliasName: aws.String(name),
        BotId:        aws.String(d.Get("bot_id").(string)),
        BotVersion:   aws.String(d.Get("bot_version").(string)),
    }

    if v, ok := d.GetOk("description"); ok {
        input.Description = aws.String(v.(string))
    }

    defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
    tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

    if len(tags) > 0 {
        input.Tags = Tags(tags.IgnoreAWS())
    }

    output, err := conn.CreateBotAlias(ctx, input)
    if err != nil {
        return diag.Errorf("error creating Lex V2 Bot Alias (%s): %s", name, err)
    }

    d.SetId(fmt.Sprintf("%s:%s", aws.ToString(output.BotAliasId), d.Get("bot_id").(string)))

    if _, err := waitBotAliasCreated(ctx, conn, aws.ToString(output.BotAliasId), d.Get("bot_id").(string), d.Timeout(schema.TimeoutCreate)); err != nil {
        return diag.Errorf("error waiting for Lex V2 Bot Alias (%s) create: %s", d.Id(), err)
    }

    return resourceAwsLexV2ModelsBotAliasRead(ctx, d, meta)
}

func resourceAwsLexV2ModelsBotAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    conn := meta.(*conns.AWSClient).LexV2ModelsConn(ctx)

    botAliasId, botId, err := BotAliasParseID(d.Id())
    if err != nil {
        return diag.FromErr(err)
    }

    resp, err := FindBotAliasByID(ctx, conn, botAliasId, botId)

    if !d.IsNewResource() && tfresource.NotFound(err) {
        log.Printf("[WARN] Lex V2 Bot Alias (%s) not found, removing from state", d.Id())
        d.SetId("")
        return nil
    }

    if err != nil {
        return diag.Errorf("error reading Lex V2 Bot Alias (%s): %s", d.Id(), err)
    }

    d.Set("bot_alias_id", resp.BotAliasId)
    d.Set("bot_alias_status", resp.BotAliasStatus)
    d.Set("bot_id", botId)
    d.Set("bot_version", resp.BotVersion)
    d.Set("description", resp.Description)
    d.Set("name", resp.BotAliasName)

    arn := arn.ARN{
        Partition: meta.(*conns.AWSClient).Partition,
        Service:   "lex",
        Region:    meta.(*conns.AWSClient).Region,
        AccountID: meta.(*conns.AWSClient).AccountID,
        Resource:  fmt.Sprintf("bot-alias/%s", aws.ToString(resp.BotAliasId)),
    }.String()
    d.Set("arn", arn)

    return nil
}

func resourceAwsLexV2ModelsBotAliasUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    conn := meta.(*conns.AWSClient).LexV2ModelsConn(ctx)

    botAliasId, botId, err := BotAliasParseID(d.Id())
    if err != nil {
        return diag.FromErr(err)
    }

    input := &lexmodelsv2.UpdateBotAliasInput{
        BotAliasId: aws.String(botAliasId),
        BotId:      aws.String(botId),
        BotVersion: aws.String(d.Get("bot_version").(string)),
    }

    if d.HasChange("description") {
        input.Description = aws.String(d.Get("description").(string))
    }

    _, err = conn.UpdateBotAlias(ctx, input)
    if err != nil {
        return diag.Errorf("error updating Lex V2 Bot Alias (%s): %s", d.Id(), err)
    }

    if d.HasChange("tags_all") {
        o, n := d.GetChange("tags_all")
        if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
            return diag.Errorf("error updating tags for Lex V2 Bot Alias (%s): %s", d.Id(), err)
        }
    }

    if _, err := waitBotAliasUpdated(ctx, conn, botAliasId, botId, d.Timeout(schema.TimeoutUpdate)); err != nil {
        return diag.Errorf("error waiting for Lex V2 Bot Alias (%s) update: %s", d.Id(), err)
    }

    return resourceAwsLexV2ModelsBotAliasRead(ctx, d, meta)
}

func resourceAwsLexV2ModelsBotAliasDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    conn := meta.(*conns.AWSClient).LexV2ModelsConn(ctx)

    botAliasId, botId, err := BotAliasParseID(d.Id())
    if err != nil {
        return diag.FromErr(err)
    }

    input := &lexmodelsv2.DeleteBotAliasInput{
        BotAliasId: aws.String(botAliasId),
        BotId:      aws.String(botId),
    }

    _, err = conn.DeleteBotAlias(ctx, input)
    if err != nil {
        if tfawserr.ErrCodeEquals(err, lexmodelsv2.ErrCodeResourceNotFoundException) {
            return nil
        }
        return diag.Errorf("error deleting Lex V2 Bot Alias (%s): %s", d.Id(), err)
    }

    if _, err := waitBotAliasDeleted(ctx, conn, botAliasId, botId, d.Timeout(schema.TimeoutDelete)); err != nil {
        return diag.Errorf("error waiting for Lex V2 Bot Alias (%s) delete: %s", d.Id(), err)
    }

    return nil
}

func resourceAwsLexV2ModelsBotAliasImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
    parts := strings.Split(d.Id(), ":")
    if len(parts) != 2 {
        return nil, fmt.Errorf("invalid import format. Expected 'bot_alias_id:bot_id', got: %s", d.Id())
    }

    d.SetId(d.Id())
    return []*schema.ResourceData{d}, nil
}