package anxcloud

import (
	"context"
	"time"

	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/anexia-it/go-anxcloud/pkg/core/resource"
	"github.com/anexia-it/go-anxcloud/pkg/core/tags"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTag() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTagCreate,
		ReadContext:   resourceTagRead,
		DeleteContext: resourceTagDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: schemaTag(),
	}
}

func resourceTagCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
	t := tags.NewAPI(c)

	def := tags.Create{
		Name:       d.Get("name").(string),
		ServiceID:  d.Get("service_id").(string),
		CustomerID: d.Get("customer_id").(string),
	}

	res, err := t.Create(ctx, def)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(res.Identifier)

	return resourceTagRead(ctx, d, m)
}

func resourceTagRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags []diag.Diagnostic

	c := m.(client.Client)
	t := tags.NewAPI(c)

	info, err := t.Get(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("organisation_assignments", flattenOrganisationAssignments(info.Organisations)); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceTagDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.Client)
	t := tags.NewAPI(c)

	if err := t.Delete(ctx, d.Id(), d.Get("service_id").(string)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func attachTag(ctx context.Context, c client.Client, resourceID, tagName string) error {
	r := resource.NewAPI(c)
	if _, err := r.AttachTag(ctx, resourceID, tagName); err != nil {
		return err
	}
	return nil
}

func detachTag(ctx context.Context, c client.Client, resourceID, tagName string) error {
	r := resource.NewAPI(c)
	if err := r.DetachTag(ctx, resourceID, tagName); err != nil {
		return err
	}
	return nil
}