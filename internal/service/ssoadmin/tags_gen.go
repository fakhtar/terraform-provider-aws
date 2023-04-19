// Code generated by internal/generate/tags/main.go; DO NOT EDIT.
package ssoadmin

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/aws/aws-sdk-go/service/ssoadmin/ssoadminiface"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// ListTags lists ssoadmin service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func ListTags(ctx context.Context, conn ssoadminiface.SSOAdminAPI, identifier, resourceType string) (tftags.KeyValueTags, error) {
	input := &ssoadmin.ListTagsForResourceInput{
		ResourceArn: aws.String(identifier),
		InstanceArn: aws.String(resourceType),
	}

	output, err := conn.ListTagsForResourceWithContext(ctx, input)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return KeyValueTags(ctx, output.Tags), nil
}

// ListTags lists ssoadmin service tags and set them in Context.
// It is called from outside this package.
func (p *servicePackage) ListTags(ctx context.Context, meta any, identifier, resourceType string) error {
	tags, err := ListTags(ctx, meta.(*conns.AWSClient).SSOAdminConn(), identifier, resourceType)

	if err != nil {
		return err
	}

	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = types.Some(tags)
	}

	return nil
}

// []*SERVICE.Tag handling

// Tags returns ssoadmin service tags.
func Tags(tags tftags.KeyValueTags) []*ssoadmin.Tag {
	result := make([]*ssoadmin.Tag, 0, len(tags))

	for k, v := range tags.Map() {
		tag := &ssoadmin.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		result = append(result, tag)
	}

	return result
}

// KeyValueTags creates tftags.KeyValueTags from ssoadmin service tags.
func KeyValueTags(ctx context.Context, tags []*ssoadmin.Tag) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.StringValue(tag.Key)] = tag.Value
	}

	return tftags.New(ctx, m)
}

// GetTagsIn returns ssoadmin service tags from Context.
// nil is returned if there are no input tags.
func GetTagsIn(ctx context.Context) []*ssoadmin.Tag {
	if inContext, ok := tftags.FromContext(ctx); ok {
		if tags := Tags(inContext.TagsIn.UnwrapOrDefault()); len(tags) > 0 {
			return tags
		}
	}

	return nil
}

// SetTagsOut sets ssoadmin service tags in Context.
func SetTagsOut(ctx context.Context, tags []*ssoadmin.Tag) {
	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = types.Some(KeyValueTags(ctx, tags))
	}
}

// UpdateTags updates ssoadmin service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func UpdateTags(ctx context.Context, conn ssoadminiface.SSOAdminAPI, identifier, resourceType string, oldTagsMap, newTagsMap any) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &ssoadmin.UntagResourceInput{
			ResourceArn: aws.String(identifier),
			InstanceArn: aws.String(resourceType),
			TagKeys:     aws.StringSlice(removedTags.IgnoreSystem(names.SSOAdmin).Keys()),
		}

		_, err := conn.UntagResourceWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &ssoadmin.TagResourceInput{
			ResourceArn: aws.String(identifier),
			InstanceArn: aws.String(resourceType),
			Tags:        Tags(updatedTags.IgnoreSystem(names.SSOAdmin)),
		}

		_, err := conn.TagResourceWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// UpdateTags updates ssoadmin service tags.
// It is called from outside this package.
func (p *servicePackage) UpdateTags(ctx context.Context, meta any, identifier, resourceType string, oldTags, newTags any) error {
	return UpdateTags(ctx, meta.(*conns.AWSClient).SSOAdminConn(), identifier, resourceType, oldTags, newTags)
}
