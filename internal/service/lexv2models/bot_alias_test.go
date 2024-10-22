package lexmodelsv2

import (
    "context"
    "fmt"
    "testing"

    "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
    "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
    "github.com/hashicorp/terraform-provider-aws/internal/acctest"
    "github.com/hashicorp/terraform-provider-aws/internal/conns"
    "github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// Basic CRUD Test
func TestAccLexV2ModelsBot_basic(t *testing.T) {
    ctx := context.Background()
    var botAlias lexmodelsv2.DescribeBotAliasOutput
    rName := fmt.Sprintf("tf-test-bot-%s", acctest.RandString(8))
    resourceName := "aws_lexv2models_bot_alias.test"

    resource.ParallelTest(t, resource.TestCase{
        PreCheck: func() {
            acctest.PreCheck(t)
            acctest.PreCheckPartitionHasService(lexmodelsv2.EndpointsID, t)
        },
        ErrorCheck:               acctest.ErrorCheck(t, lexmodelsv2.EndpointsID),
        ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
        CheckDestroy:             testAccCheckBotAliasDestroy(ctx),
        Steps: []resource.TestStep{
            {
                Config: testAccBotAliasConfig_basic(rName),
                Check: resource.ComposeTestCheckFunc(
                    testAccCheckBotAliasExists(ctx, resourceName, &botAlias),
                    resource.TestCheckResourceAttr(resourceName, "name", rName),
                    resource.TestCheckResourceAttr(resourceName, "description", "Test bot alias"),
                    resource.TestCheckResourceAttr(resourceName, "bot_version", "1"),
                    acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lex", fmt.Sprintf("bot-alias/%s", rName)),
                ),
            },
            {
                Config: testAccBotAliasConfig_updated(rName),
                Check: resource.ComposeTestCheckFunc(
                    testAccCheckBotAliasExists(ctx, resourceName, &botAlias),
                    resource.TestCheckResourceAttr(resourceName, "name", rName),
                    resource.TestCheckResourceAttr(resourceName, "description", "Updated test bot alias"),
                    resource.TestCheckResourceAttr(resourceName, "bot_version", "2"),
                ),
            },
        },
    })
}

// Import Test
func TestAccLexV2ModelsBot_import(t *testing.T) {
    ctx := context.Background()
    var botAlias lexmodelsv2.DescribeBotAliasOutput
    rName := fmt.Sprintf("tf-test-bot-%s", acctest.RandString(8))
    resourceName := "aws_lexv2models_bot_alias.test"

    resource.ParallelTest(t, resource.TestCase{
        PreCheck: func() {
            acctest.PreCheck(t)
            acctest.PreCheckPartitionHasService(lexmodelsv2.EndpointsID, t)
        },
        ErrorCheck:               acctest.ErrorCheck(t, lexmodelsv2.EndpointsID),
        ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
        CheckDestroy:             testAccCheckBotAliasDestroy(ctx),
        Steps: []resource.TestStep{
            {
                Config: testAccBotAliasConfig_basic(rName),
                Check: resource.ComposeTestCheckFunc(
                    testAccCheckBotAliasExists(ctx, resourceName, &botAlias),
                ),
            },
            {
                ResourceName:      resourceName,
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}

// Disappears Test
func TestAccLexV2ModelsBot_disappears(t *testing.T) {
    ctx := context.Background()
    var botAlias lexmodelsv2.DescribeBotAliasOutput
    rName := fmt.Sprintf("tf-test-bot-%s", acctest.RandString(8))
    resourceName := "aws_lexv2models_bot_alias.test"

    resource.ParallelTest(t, resource.TestCase{
        PreCheck: func() {
            acctest.PreCheck(t)
            acctest.PreCheckPartitionHasService(lexmodelsv2.EndpointsID, t)
        },
        ErrorCheck:               acctest.ErrorCheck(t, lexmodelsv2.EndpointsID),
        ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
        CheckDestroy:             testAccCheckBotAliasDestroy(ctx),
        Steps: []resource.TestStep{
            {
                Config: testAccBotAliasConfig_basic(rName),
                Check: resource.ComposeTestCheckFunc(
                    testAccCheckBotAliasExists(ctx, resourceName, &botAlias),
                    acctest.CheckResourceDisappears(ctx, acctest.Provider, resourceAwsLexV2ModelsBot(), resourceName),
                ),
                ExpectNonEmptyPlan: true,
            },
        },
    })
}

// Tags Test
func TestAccLexV2ModelsBot_tags(t *testing.T) {
    ctx := context.Background()
    var botAlias lexmodelsv2.DescribeBotAliasOutput
    rName := fmt.Sprintf("tf-test-bot-%s", acctest.RandString(8))
    resourceName := "aws_lexv2models_bot_alias.test"

    resource.ParallelTest(t, resource.TestCase{
        PreCheck: func() {
            acctest.PreCheck(t)
            acctest.PreCheckPartitionHasService(lexmodelsv2.EndpointsID, t)
        },
        ErrorCheck:               acctest.ErrorCheck(t, lexmodelsv2.EndpointsID),
        ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
        CheckDestroy:             testAccCheckBotAliasDestroy(ctx),
        Steps: []resource.TestStep{
            {
                Config: testAccBotAliasConfig_tags1(rName, "key1", "value1"),
                Check: resource.ComposeTestCheckFunc(
                    testAccCheckBotAliasExists(ctx, resourceName, &botAlias),
                    resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
                    resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
                ),
            },
            {
                Config: testAccBotAliasConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
                Check: resource.ComposeTestCheckFunc(
                    testAccCheckBotAliasExists(ctx, resourceName, &botAlias),
                    resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
                    resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
                    resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
                ),
            },
        },
    })
}

// Helper Functions
func testAccCheckBotAliasDestroy(ctx context.Context) resource.TestCheckFunc {
    return func(s *terraform.State) error {
        conn := acctest.Provider.Meta().(*conns.AWSClient).LexV2ModelsConn(ctx)

        for _, rs := range s.RootModule().Resources {
            if rs.Type != "aws_lexv2models_bot_alias" {
                continue
            }

            botAliasId, botId, err := BotAliasParseID(rs.Primary.ID)
            if err != nil {
                return err
            }

            _, err = FindBotAliasByID(ctx, conn, botAliasId, botId)

            if tfresource.NotFound(err) {
                continue
            }

            if err != nil {
                return err
            }

            return fmt.Errorf("Lex V2 Bot Alias %s still exists", rs.Primary.ID)
        }

        return nil
    }
}

func testAccCheckBotAliasExists(ctx context.Context, n string, v *lexmodelsv2.DescribeBotAliasOutput) resource.TestCheckFunc {
    return func(s *terraform.State) error {
        rs, ok := s.RootModule().Resources[n]
        if !ok {
            return fmt.Errorf("Not found: %s", n)
        }

        if rs.Primary.ID == "" {
            return fmt.Errorf("No Lex V2 Bot Alias ID is set")
        }

        conn := acctest.Provider.Meta().(*conns.AWSClient).LexV2ModelsConn(ctx)

        botAliasId, botId, err := BotAliasParseID(rs.Primary.ID)
        if err != nil {
            return err
        }

        resp, err := FindBotAliasByID(ctx, conn, botAliasId, botId)

        if err != nil {
            return err
        }

        *v = *resp

        return nil
    }
}

// Test Configurations
func testAccBotAliasConfig_basic(rName string) string {
    return fmt.Sprintf(`
resource "aws_lexv2models_bot" "test" {
  name = %[1]q
  data_privacy {
    child_directed = false
  }
  idle_session_ttl_in_seconds = 300
}

resource "aws_lexv2models_bot_version" "test" {
  bot_id = aws_lexv2models_bot.test.id
}

resource "aws_lexv2models_bot_alias" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_version.test.bot_version
  name        = %[1]q
  description = "Test bot alias"
}
`, rName)
}

func testAccBotAliasConfig_updated(rName string) string {
    return fmt.Sprintf(`
resource "aws_lexv2models_bot" "test" {
  name = %[1]q
  data_privacy {
    child_directed = false
  }
  idle_session_ttl_in_seconds = 300
}

resource "aws_lexv2models_bot_version" "test" {
  bot_id = aws_lexv2models_bot.test.id
}

resource "aws_lexv2models_bot_version" "test2" {
  bot_id = aws_lexv2models_bot.test.id
}

resource "aws_lexv2models_bot_alias" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_version.test2.bot_version
  name        = %[1]q
  description = "Updated test bot alias"
}
`, rName)
}

func testAccBotAliasConfig_tags1(rName, tagKey1, tagValue1 string) string {
    return fmt.Sprintf(`
resource "aws_lexv2models_bot" "test" {
  name = %[1]q
  data_privacy {
    child_directed = false
  }
  idle_session_ttl_in_seconds = 300
}

resource "aws_lexv2models_bot_version" "test" {
  bot_id = aws_lexv2models_bot.test.id
}

resource "aws_lexv2models_bot_alias" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_version.test.bot_version
  name        = %[1]q
  description = "Test bot alias"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccBotAliasConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
    return fmt.Sprintf(`
resource "aws_lexv2models_bot" "test" {
  name = %[1]q
  data_privacy {
    child_directed = false
  }
  idle_session_ttl_in_seconds = 300
}

resource "aws_lexv2models_bot_version" "test" {
  bot_id = aws_lexv2models_bot.test.id
}

resource "aws_lexv2models_bot_alias" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_version.test.bot_version
  name        = %[1]q
  description = "Test bot alias"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}