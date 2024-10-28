// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
// @Tags(identifierAttribute="arn")
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
			},
			"bot_alias_status": schema.StringAttribute{
				Computed: true,
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
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
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
		BotAliasName: aws.String(plan.Name.ValueString()),
		BotId:        aws.String(plan.BotID.ValueString()),
		BotVersion:   aws.String(plan.BotVersion.ValueString()),
		Tags:         getTagsIn(ctx),
	}

	if !plan.Description.IsNull() {
		in.Description = aws.String(plan.Description.ValueString())
	}

	out, err := conn.CreateBotAlias(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameBotAlias, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameBotAlias, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	id := fmt.Sprintf("%s:%s", aws.ToString(out.BotAliasId), plan.BotID.ValueString())
	plan.ID = flex.StringToFramework(ctx, &id)
	plan.BotAliasID = flex.StringToFramework(ctx, out.BotAliasId)

	botAliasArn := arn.ARN{
		Partition: r.Meta().Partition,
		Service:   "lex",
		Region:    r.Meta().Region,
		AccountID: r.Meta().AccountID,
		Resource:  fmt.Sprintf("bot-alias/%s", aws.ToString(out.BotAliasId)),
	}.String()
	plan.ARN = flex.StringToFramework(ctx, &botAliasArn)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitBotAliasCreated(ctx, conn, plan.BotAliasID.ValueString(), plan.BotID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForCreation, ResNameBotAlias, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceBotAlias) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var state resourceBotAliasData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	botAliasId, botId, err := BotAliasParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionReading, ResNameBotAlias, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	out, err := FindBotAliasByID(ctx, conn, botAliasId, botId)
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

	state.BotID = flex.StringToFramework(ctx, &botId)
	diags := state.refreshFromOutput(ctx, out)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceBotAlias) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var plan, state resourceBotAliasData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) ||
		!plan.BotVersion.Equal(state.BotVersion) {

		botAliasId, botId, err := BotAliasParseID(state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LexV2Models, create.ErrActionUpdating, ResNameBotAlias, state.ID.String(), err),
				err.Error(),
			)
			return
		}

		in := &lexmodelsv2.UpdateBotAliasInput{
			BotAliasId: aws.String(botAliasId),
			BotId:      aws.String(botId),
			BotVersion: aws.String(plan.BotVersion.ValueString()),
		}

		if !plan.Description.IsNull() {
			in.Description = aws.String(plan.Description.ValueString())
		}

		_, err = conn.UpdateBotAlias(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LexV2Models, create.ErrActionUpdating, ResNameBotAlias, state.ID.String(), err),
				err.Error(),
			)
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		_, err = waitBotAliasUpdated(ctx, conn, botAliasId, botId, updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForUpdate, ResNameBotAlias, state.ID.String(), err),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceBotAlias) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var state resourceBotAliasData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	botAliasId, botId, err := BotAliasParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionDeleting, ResNameBotAlias, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	in := &lexmodelsv2.DeleteBotAliasInput{
		BotAliasId: aws.String(botAliasId),
		BotId:      aws.String(botId),
	}

	_, err = conn.DeleteBotAlias(ctx, in)
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
	_, err = waitBotAliasDeleted(ctx, conn, botAliasId, botId, deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForDeletion, ResNameBotAlias, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceBotAlias) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func (r *resourceBotAlias) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

type resourceBotAliasData struct {
	ARN            types.String   `tfsdk:"arn"`
	BotAliasID     types.String   `tfsdk:"bot_alias_id"`
	BotAliasStatus types.String   `tfsdk:"bot_alias_status"`
	BotID          types.String   `tfsdk:"bot_id"`
	BotVersion     types.String   `tfsdk:"bot_version"`
	Description    types.String   `tfsdk:"description"`
	ID             types.String   `tfsdk:"id"`
	Name           types.String   `tfsdk:"name"`
	Tags           tftags.Map     `tfsdk:"tags"`
	TagsAll        tftags.Map     `tfsdk:"tags_all"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}

func (rd *resourceBotAliasData) refreshFromOutput(ctx context.Context, out *lexmodelsv2.DescribeBotAliasOutput) diag.Diagnostics {
	var diags diag.Diagnostics

	if out == nil {
		return diags
	}

	rd.BotAliasID = flex.StringToFramework(ctx, out.BotAliasId)
	rd.BotAliasStatus = flex.StringToFramework(ctx, (*string)(&out.BotAliasStatus))
	rd.BotVersion = flex.StringToFramework(ctx, out.BotVersion)
	rd.Description = flex.StringToFramework(ctx, out.Description)
	rd.Name = flex.StringToFramework(ctx, out.BotAliasName)

	return diags
}
func waitBotAliasCreated(ctx context.Context, conn *lexmodelsv2.Client, botAliasId, botId string, timeout time.Duration) (*lexmodelsv2.DescribeBotAliasOutput, error) {
    stateConf := &retry.StateChangeConf{
        Pending: []string{string(awstypes.BotStatusCreating)},
        Target:  []string{string(awstypes.BotStatusAvailable)},
        Refresh: statusBotAlias(ctx, conn, botAliasId, botId),
        Timeout: timeout,
    }

    outputRaw, err := stateConf.WaitForStateContext(ctx)
    if out, ok := outputRaw.(*lexmodelsv2.DescribeBotAliasOutput); ok {
        return out, err
    }

    return nil, err
}

func waitBotAliasUpdated(ctx context.Context, conn *lexmodelsv2.Client, botAliasId, botId string, timeout time.Duration) (*lexmodelsv2.DescribeBotAliasOutput, error) {
    stateConf := &retry.StateChangeConf{
        Pending: []string{string(awstypes.BotStatusUpdating)},
        Target:  []string{string(awstypes.BotStatusAvailable)},
        Refresh: statusBotAlias(ctx, conn, botAliasId, botId),
        Timeout: timeout,
    }

    outputRaw, err := stateConf.WaitForStateContext(ctx)
    if out, ok := outputRaw.(*lexmodelsv2.DescribeBotAliasOutput); ok {
        return out, err
    }

    return nil, err
}

func waitBotAliasDeleted(ctx context.Context, conn *lexmodelsv2.Client, botAliasId, botId string, timeout time.Duration) (*lexmodelsv2.DescribeBotAliasOutput, error) {
    stateConf := &retry.StateChangeConf{
        Pending: []string{string(awstypes.BotStatusDeleting)},
        Target:  []string{},
        Refresh: statusBotAlias(ctx, conn, botAliasId, botId),
        Timeout: timeout,
    }

    outputRaw, err := stateConf.WaitForStateContext(ctx)
    if out, ok := outputRaw.(*lexmodelsv2.DescribeBotAliasOutput); ok {
        return out, err
    }

    return nil, err
}

func statusBotAlias(ctx context.Context, conn *lexmodelsv2.Client, botAliasId, botId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindBotAliasByID(ctx, conn, botAliasId, botId)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.BotAliasStatus), nil
	}
}

func FindBotAliasByID(ctx context.Context, conn *lexmodelsv2.Client, botAliasId, botId string) (*lexmodelsv2.DescribeBotAliasOutput, error) {
	in := &lexmodelsv2.DescribeBotAliasInput{
		BotAliasId: aws.String(botAliasId),
		BotId:      aws.String(botId),
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

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func BotAliasParseID(id string) (botAliasId, botId string, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = fmt.Errorf("unexpected format of ID (%s), expected BOT-ALIAS-ID:BOT-ID", id)
		return
	}

	botAliasId = parts[0]
	botId = parts[1]
	return
}