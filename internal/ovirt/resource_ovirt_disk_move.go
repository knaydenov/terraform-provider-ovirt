package ovirt

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

var diskMoveSchema = map[string]*schema.Schema{
	"disk_id": {
		Type:             schema.TypeString,
		Required:         true,
		ForceNew:         true,
		Description:      "ID of the disk to resize.",
		ValidateDiagFunc: validateUUID,
	},
	"storage_domain_id": {
		Type:             schema.TypeString,
		Required:         true,
		Description:      "ID of the target storage domain.",
		ForceNew:         true,
		ValidateDiagFunc: validateUUID,
	},
}

func (p *provider) diskMoveResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: p.diskMoveCreate,
		ReadContext:   p.diskMoveRead,
		DeleteContext: p.diskMoveDelete,
		Schema:        diskMoveSchema,
		Description:   `The ovirt_disk_move resource moves disks in oVirt to the specified storage domain.`,
	}
}

func (p *provider) diskMoveCreate(
	ctx context.Context,
	data *schema.ResourceData,
	_ interface{},
) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	return moveDisk(client, data)
}

func (p *provider) diskMoveRead(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)

	diskID := data.Get("disk_id").(string)
	targetStorageDomainID := data.Get("target_storage_domain_id").(string)

	disk, err := client.GetDisk(ovirtclient.DiskID(diskID))
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve disk.",
				Detail:   err.Error(),
			},
		}
	}

	targetStorageDomain, err := client.GetStorageDomain(ovirtclient.StorageDomainID(targetStorageDomainID))
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve target storage domain.",
				Detail:   err.Error(),
			},
		}
	}

	data.SetId(string(disk.ID()))
	diags := diag.Diagnostics{}
	diags = setResourceField(data, "storage_domain_id", string(targetStorageDomain.ID()), diags)

	return diags
}

func (p *provider) diskMoveDelete(_ context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	data.SetId("")
	return nil
}

func moveDisk(client ovirtclient.Client, data *schema.ResourceData) diag.Diagnostics {
	diskID := data.Get("disk_id").(string)
	targetStorageDomainID := data.Get("target_storage_domain_id").(string)

	disk, err := client.GetDisk(ovirtclient.DiskID(diskID))
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve disk.",
				Detail:   err.Error(),
			},
		}
	}

	targetStorageDomain, err := client.GetStorageDomain(ovirtclient.StorageDomainID(targetStorageDomainID))
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve target storage domain.",
				Detail:   err.Error(),
			},
		}
	}

	disk, err = client.MoveDiskToStorageDomain(disk.ID(), targetStorageDomain.ID())
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to move disk.",
				Detail:   err.Error(),
			},
		}
	}

	data.SetId(string(disk.ID()))
	diags := diag.Diagnostics{}
	diags = setResourceField(data, "target_storage_domain_id", string(targetStorageDomain.ID()), diags)

	return diags
}
