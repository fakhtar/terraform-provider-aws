// Code generated by internal/generate/tags/main.go; DO NOT EDIT.
package firehose

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/firehose/firehoseiface"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// ListTags lists firehose service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func ListTags(ctx context.Context, conn firehoseiface.FirehoseAPI, identifier string) (tftags.KeyValueTags, error) {
	input := &firehose.ListTagsForDeliveryStreamInput{
		DeliveryStreamName: aws.String(identifier),
	}

	output, err := conn.ListTagsForDeliveryStreamWithContext(ctx, input)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return KeyValueTags(ctx, output.Tags), nil
}

// ListTags lists firehose service tags and set them in Context.
// It is called from outside this package.
func (p *servicePackage) ListTags(ctx context.Context, meta any, identifier string) error {
	tags, err := ListTags(ctx, meta.(*conns.AWSClient).FirehoseConn(), identifier)

	if err != nil {
		return err
	}

	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = types.Some(tags)
	}

	return nil
}

// []*SERVICE.Tag handling

// Tags returns firehose service tags.
func Tags(tags tftags.KeyValueTags) []*firehose.Tag {
	result := make([]*firehose.Tag, 0, len(tags))

	for k, v := range tags.Map() {
		tag := &firehose.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		result = append(result, tag)
	}

	return result
}

// KeyValueTags creates tftags.KeyValueTags from firehose service tags.
func KeyValueTags(ctx context.Context, tags []*firehose.Tag) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.StringValue(tag.Key)] = tag.Value
	}

	return tftags.New(ctx, m)
}

// GetTagsIn returns firehose service tags from Context.
// nil is returned if there are no input tags.
func GetTagsIn(ctx context.Context) []*firehose.Tag {
	if inContext, ok := tftags.FromContext(ctx); ok {
		if tags := Tags(inContext.TagsIn.UnwrapOrDefault()); len(tags) > 0 {
			return tags
		}
	}

	return nil
}

// SetTagsOut sets firehose service tags in Context.
func SetTagsOut(ctx context.Context, tags []*firehose.Tag) {
	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = types.Some(KeyValueTags(ctx, tags))
	}
}

// UpdateTags updates firehose service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func UpdateTags(ctx context.Context, conn firehoseiface.FirehoseAPI, identifier string, oldTagsMap, newTagsMap any) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &firehose.UntagDeliveryStreamInput{
			DeliveryStreamName: aws.String(identifier),
			TagKeys:            aws.StringSlice(removedTags.IgnoreSystem(names.Firehose).Keys()),
		}

		_, err := conn.UntagDeliveryStreamWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &firehose.TagDeliveryStreamInput{
			DeliveryStreamName: aws.String(identifier),
			Tags:               Tags(updatedTags.IgnoreSystem(names.Firehose)),
		}

		_, err := conn.TagDeliveryStreamWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// UpdateTags updates firehose service tags.
// It is called from outside this package.
func (p *servicePackage) UpdateTags(ctx context.Context, meta any, identifier string, oldTags, newTags any) error {
	return UpdateTags(ctx, meta.(*conns.AWSClient).FirehoseConn(), identifier, oldTags, newTags)
}
