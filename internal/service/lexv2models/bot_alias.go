package lexv2models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Bot Alias")
func newResourceBotAlias(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceBotAlias{}
	
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameBotAlias = "Bot Alias"
)

type resourceBotAlias struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceBotAlias) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_lexv2models_bot_alias"
}

func (r *resourceBotAlias) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"bot_alias_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bot_alias_name": schema.StringAttribute{
				Required: true,
			},
			"bot_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bot_version": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"sentiment_analysis_settings": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"detect_sentiment": schema.BoolAttribute{
							Required: true,
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceBotAlias) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var plan resourceBotAliasData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lexmodelsv2.CreateBotAliasInput{
		BotAliasName: aws.String(plan.BotAliasName.ValueString()),
		BotId:        aws.String(plan.BotId.ValueString()),
		BotVersion:   aws.String(plan.BotVersion.ValueString()),
		Tags:         getTagsIn(ctx),
	}

	if !plan.Description.IsNull() {
		in.Description = aws.String(plan.Description.ValueString())
	}

	if !plan.SentimentAnalysisSettings.IsNull() {
		var tfList []sentimentAnalysisSettingsData
		resp.Diagnostics.Append(plan.SentimentAnalysisSettings.ElementsAs(ctx, &tfList, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.SentimentAnalysisSettings = expandSentimentAnalysisSettings(ctx, tfList)
	}

	out, err := conn.CreateBotAlias(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameBotAlias, plan.BotAliasName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.BotAliasId == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameBotAlias, plan.BotAliasName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = types.StringValue(aws.ToString(out.BotAliasId))
	plan.BotAliasId = types.StringValue(aws.ToString(out.BotAliasId))

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitBotAliasCreated(ctx, conn, plan.BotId.ValueString(), plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForCreation, ResNameBotAlias, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	// Set computed values
	botArn := arn.ARN{
		Partition: r.Meta().Partition,
		Service:   "lex",
		Region:    r.Meta().Region,
		AccountID: r.Meta().AccountID,
		Resource:  fmt.Sprintf("bot-alias/%s", aws.ToString(out.BotAliasId)),
	}.String()
	plan.ARN = types.StringValue(botArn)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Other CRUD operations and helper functions will follow...

type resourceBotAliasData struct {
	ARN                       types.String   `tfsdk:"arn"`
	BotAliasId               types.String   `tfsdk:"bot_alias_id"`
	BotAliasName             types.String   `tfsdk:"bot_alias_name"`
	BotId                    types.String   `tfsdk:"bot_id"`
	BotVersion               types.String   `tfsdk:"bot_version"`
	Description              types.String   `tfsdk:"description"`
	ID                       types.String   `tfsdk:"id"`
	SentimentAnalysisSettings types.List     `tfsdk:"sentiment_analysis_settings"`
	Tags                     types.Map      `tfsdk:"tags"`
	TagsAll                  types.Map      `tfsdk:"tags_all"`
	Timeouts                 timeouts.Value `tfsdk:"timeouts"`
}

type sentimentAnalysisSettingsData struct {
	DetectSentiment types.Bool `tfsdk:"detect_sentiment"`
}

func expandSentimentAnalysisSettings(ctx context.Context, tfList []sentimentAnalysisSettingsData) *awstypes.SentimentAnalysisSettings {
	if len(tfList) == 0 {
		return nil
	}

	return &awstypes.SentimentAnalysisSettings{
		DetectSentiment: tfList[0].DetectSentiment.ValueBool(),
	}
}

// Read operation implementation
func (r *resourceBotAlias) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)
	var state resourceBotAliasData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindBotAliasByID(ctx, conn, state.BotId.ValueString(), state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionReading, ResNameBotAlias, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.BotAliasName = flex.StringToFramework(ctx, out.BotAliasName)
	state.BotVersion = flex.StringToFramework(ctx, out.BotVersion)
	state.Description = flex.StringToFramework(ctx, out.Description)

	botArn := arn.ARN{
		Partition: r.Meta().Partition,
		Service:   "lex",
		Region:    r.Meta().Region,
		AccountID: r.Meta().AccountID,
		Resource:  fmt.Sprintf("bot-alias/%s", aws.ToString(out.BotAliasId)),
	}.String()
	state.ARN = types.StringValue(botArn)

	if out.SentimentAnalysisSettings != nil {
		settings, d := flattenSentimentAnalysisSettings(ctx, out.SentimentAnalysisSettings)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.SentimentAnalysisSettings = settings
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update operation implementation
func (r *resourceBotAlias) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var plan, state resourceBotAliasData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.BotAliasName.Equal(state.BotAliasName) ||
		!plan.BotVersion.Equal(state.BotVersion) ||
		!plan.Description.Equal(state.Description) ||
		!plan.SentimentAnalysisSettings.Equal(state.SentimentAnalysisSettings) {

		in := &lexmodelsv2.UpdateBotAliasInput{
			BotAliasId:   aws.String(plan.ID.ValueString()),
			BotId:        aws.String(plan.BotId.ValueString()),
			BotAliasName: aws.String(plan.BotAliasName.ValueString()),
			BotVersion:   aws.String(plan.BotVersion.ValueString()),
		}

		if !plan.Description.IsNull() {
			in.Description = aws.String(plan.Description.ValueString())
		}

		if !plan.SentimentAnalysisSettings.IsNull() {
			var tfList []sentimentAnalysisSettingsData
			resp.Diagnostics.Append(plan.SentimentAnalysisSettings.ElementsAs(ctx, &tfList, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			in.SentimentAnalysisSettings = expandSentimentAnalysisSettings(ctx, tfList)
		}

		_, err := conn.UpdateBotAlias(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LexV2Models, create.ErrActionUpdating, ResNameBotAlias, plan.ID.String(), err),
				err.Error(),
			)
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		_, err = waitBotAliasUpdated(ctx, conn, plan.BotId.ValueString(), plan.ID.ValueString(), updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForUpdate, ResNameBotAlias, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete operation implementation
func (r *resourceBotAlias) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var state resourceBotAliasData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lexmodelsv2.DeleteBotAliasInput{
		BotAliasId: aws.String(state.ID.ValueString()),
		BotId:      aws.String(state.BotId.ValueString()),
	}

	_, err := conn.DeleteBotAlias(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionDeleting, ResNameBotAlias, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitBotAliasDeleted(ctx, conn, state.BotId.ValueString(), state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForDeletion, ResNameBotAlias, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

// Waiter implementations
const (
	BotAliasStatusCreating  = "Creating"
	BotAliasStatusUpdating  = "Updating"
	BotAliasStatusAvailable = "Available"
	BotAliasStatusDeleting  = "Deleting"
)

func waitBotAliasCreated(ctx context.Context, conn *lexmodelsv2.Client, botID, botAliasID string, timeout time.Duration) (*lexmodelsv2.DescribeBotAliasOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{BotAliasStatusCreating},
		Target:                    []string{BotAliasStatusAvailable},
		Refresh:                   statusBotAlias(ctx, conn, botID, botAliasID),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lexmodelsv2.DescribeBotAliasOutput); ok {
		return out, err
	}

	return nil, err
}

func waitBotAliasUpdated(ctx context.Context, conn *lexmodelsv2.Client, botID, botAliasID string, timeout time.Duration) (*lexmodelsv2.DescribeBotAliasOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{BotAliasStatusUpdating},
		Target:                    []string{BotAliasStatusAvailable},
		Refresh:                   statusBotAlias(ctx, conn, botID, botAliasID),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lexmodelsv2.DescribeBotAliasOutput); ok {
		return out, err
	}

	return nil, err
}

func waitBotAliasDeleted(ctx context.Context, conn *lexmodelsv2.Client, botID, botAliasID string, timeout time.Duration) (*lexmodelsv2.DescribeBotAliasOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{BotAliasStatusDeleting},
		Target:  []string{},
		Refresh: statusBotAlias(ctx, conn, botID, botAliasID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lexmodelsv2.DescribeBotAliasOutput); ok {
		return out, err
	}

	return nil, err
}

func statusBotAlias(ctx context.Context, conn *lexmodelsv2.Client, botID, botAliasID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindBotAliasByID(ctx, conn, botID, botAliasID)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.BotAliasStatus), nil
	}
}

func FindBotAliasByID(ctx context.Context, conn *lexmodelsv2.Client, botID, botAliasID string) (*lexmodelsv2.DescribeBotAliasOutput, error) {
	in := &lexmodelsv2.DescribeBotAliasInput{
		BotAliasId: aws.String(botAliasID),
		BotId:      aws.String(botID),
	}

	out, err := conn.DescribeBotAlias(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.BotAliasId == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenSentimentAnalysisSettings(ctx context.Context, apiObject *awstypes.SentimentAnalysisSettings) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"detect_sentiment": types.BoolType,
	}}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"detect_sentiment": types.BoolValue(apiObject.DetectSentiment),
	}
	objVal, d := types.ObjectValue(elemType.AttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}