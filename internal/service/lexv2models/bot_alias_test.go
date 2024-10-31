package lexv2models_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tflexv2models "github.com/hashicorp/terraform-provider-aws/internal/service/lexv2models"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Unit Tests
func TestWaitBotAliasCreated(t *testing.T) {
    testCases := []struct {
        name          string
        maxAttempts   int
        responses     []string
        expectedError bool
    }{
        {
            name:        "successful creation",
            maxAttempts: 3,
            responses:   []string{tflexv2models.BotAliasStatusCreating, tflexv2models.BotAliasStatusCreating, tflexv2models.BotAliasStatusAvailable},
        },
        {
            name:          "creation timeout",
            maxAttempts:   3,
            responses:     []string{tflexv2models.BotAliasStatusCreating, tflexv2models.BotAliasStatusCreating, tflexv2models.BotAliasStatusCreating},
            expectedError: true,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            ctx := context.Background()
            attempt := 0
            mockClient := &mockLexV2ModelsClientImpl{
                DescribeBotAliasFunc: func(_ context.Context, _ *lexmodelsv2.DescribeBotAliasInput, _ ...func(*lexmodelsv2.Options)) (*lexmodelsv2.DescribeBotAliasOutput, error) {
                    if attempt >= len(tc.responses) {
                        return nil, &awstypes.ResourceNotFoundException{}
                    }
                    status := tc.responses[attempt]
                    attempt++
                    return &lexmodelsv2.DescribeBotAliasOutput{
                        BotAliasStatus: awstypes.BotAliasStatus(status),
                    }, nil
                },
            }

            _, err := tflexv2models.WaitBotAliasCreated(ctx, mockClient, "testBot", "testAlias", 1*time.Second)
            if tc.expectedError && err == nil {
                t.Fatal("expected error but got none")
            }
            if !tc.expectedError && err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
        })
    }
}
func TestExpandSentimentAnalysisSettings(t *testing.T) {
    testCases := []struct {
        name     string
        input    []tflexv2models.SentimentAnalysisSettingsData
        expected awstypes.SentimentAnalysisSettings
    }{
        {
            name:  "nil input",
            input: nil,
            expected: awstypes.SentimentAnalysisSettings{
                DetectSentiment: false,
            },
        },
        {
            name: "with settings",
            input: []tflexv2models.SentimentAnalysisSettingsData{
                {
                    DetectSentiment: types.BoolValue(true),
                },
            },
            expected: awstypes.SentimentAnalysisSettings{
                DetectSentiment: true,
            },
        },
    }

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			got := tflexv2models.ExpandSentimentAnalysisSettings(ctx, tc.input)
			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

// Acceptance Tests
func TestAccLexV2ModelsAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var botAlias lexmodelsv2.DescribeBotAliasOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBotAliasConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotAliasExists(ctx, resourceName, &botAlias),
					resource.TestCheckResourceAttr(resourceName, "bot_alias_name", rName),
					resource.TestCheckResourceAttr(resourceName, "bot_version", "$LATEST"),
					resource.TestCheckResourceAttrSet(resourceName, "bot_alias_id"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
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

func TestAccLexV2ModelsAlias_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var botAlias lexmodelsv2.DescribeBotAliasOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBotAliasConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotAliasExists(ctx, resourceName, &botAlias),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflexv2models.NewResourceBotAlias, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLexV2ModelsAlias_Description(t *testing.T) {
	ctx := acctest.Context(t)
	var botAlias lexmodelsv2.DescribeBotAliasOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBotAliasConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotAliasExists(ctx, resourceName, &botAlias),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBotAliasConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotAliasExists(ctx, resourceName, &botAlias),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func testAccCheckBotAliasDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LexV2ModelsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lexv2models_bot_alias" {
				continue
			}

			_, err := tflexv2models.FindBotAliasByID(ctx, conn, rs.Primary.Attributes["bot_id"], rs.Primary.ID)
			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.LexV2Models, create.ErrActionCheckingDestroyed, tflexv2models.ResNameBotAlias, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckBotAliasExists(ctx context.Context, name string, botAlias *lexmodelsv2.DescribeBotAliasOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameBotAlias, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameBotAlias, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LexV2ModelsClient(ctx)
		resp, err := tflexv2models.FindBotAliasByID(ctx, conn, rs.Primary.Attributes["bot_id"], rs.Primary.ID)
		if err != nil {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameBotAlias, rs.Primary.ID, err)
		}

		*botAlias = *resp

		return nil
	}
}

func testAccBotAliasConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_lexv2models_bot" "test" {
  name                        = %[1]q
  description                 = "Test bot"
  idle_session_ttl_in_seconds = 60
  role_arn                    = aws_iam_role.test.arn

  data_privacy {
    child_directed = false
  }
}

resource "aws_lexv2models_bot_alias" "test" {
  bot_alias_name = %[1]q
  bot_id         = aws_lexv2models_bot.test.id
  bot_version    = "$LATEST"
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lexv2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonLexFullAccess"
}
`, rName)
}

func testAccBotAliasConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_lexv2models_bot" "test" {
  name                        = %[1]q
  description                 = "Test bot"
  idle_session_ttl_in_seconds = 60
  role_arn                    = aws_iam_role.test.arn

  data_privacy {
    child_directed = false
  }
}

resource "aws_lexv2models_bot_alias" "test" {
  bot_alias_name = %[1]q
  bot_id         = aws_lexv2models_bot.test.id
  bot_version    = "$LATEST"
  description    = %[2]q
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lexv2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonLexFullAccess"
}
`, rName, description)
}

type mockLexV2ModelsClient struct {
    lexmodelsv2.Client
}

type mockLexV2ModelsClientImpl struct {
    DescribeBotAliasFunc func(context.Context, *lexmodelsv2.DescribeBotAliasInput, ...func(*lexmodelsv2.Options)) (*lexmodelsv2.DescribeBotAliasOutput, error)
}

func (m *mockLexV2ModelsClientImpl) DescribeBotAlias(ctx context.Context, params *lexmodelsv2.DescribeBotAliasInput, optFns ...func(*lexmodelsv2.Options)) (*lexmodelsv2.DescribeBotAliasOutput, error) {
    if m.DescribeBotAliasFunc != nil {
        return m.DescribeBotAliasFunc(ctx, params, optFns...)
    }
    return nil, nil
}


func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LexV2ModelsClient(ctx)

	input := &lexmodelsv2.ListBotsInput{}
	_, err := conn.ListBots(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}