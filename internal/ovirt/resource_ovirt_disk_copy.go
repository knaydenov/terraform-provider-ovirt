package ovirt

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

var diskCopySchema = map[string]*schema.Schema{
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
		Description:      "ID of the source storage domain.",
		ForceNew:         true,
		ValidateDiagFunc: validateUUID,
	},
}

func (p *provider) diskCopyResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: p.diskCopyCreate,
		ReadContext:   p.diskCopyRead,
		DeleteContext: p.diskCopyDelete,
		Schema:        diskCopySchema,
		Description:   `The ovirt_disk_copy resource copies disk in oVirt to the specified storage domain.`,
	}
}

func (p *provider) diskCopyCreate(
	ctx context.Context,
	data *schema.ResourceData,
	_ interface{},
) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	return copyDisk(client, data)
}

func (p *provider) diskCopyRead(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	return getDisk(client, data)
}

func (p *provider) diskCopyDelete(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	return deleteDisk(client, data)
}

func copyDisk(client ovirtclient.Client, data *schema.ResourceData) diag.Diagnostics {
	diskID := data.Get("disk_id").(string)
	storageDomainID := data.Get("storage_domain_id").(string)

	templateDisk, err := client.GetDisk(ovirtclient.DiskID(diskID))
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve template disk.",
				Detail:   err.Error(),
			},
		}
	}

	storageDomain, err := client.GetStorageDomain(ovirtclient.StorageDomainID(storageDomainID))
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve target storage domain.",
				Detail:   err.Error(),
			},
		}
	}

	wait, err := client.StartCopyTemplateDiskToStorageDomain(templateDisk.ID(), storageDomain.ID())
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to copy disk.",
				Detail:   err.Error(),
			},
		}
	}

	disk, err := wait.Wait()
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to copy disk.",
				Detail:   err.Error(),
			},
		}
	}

	data.SetId(string(disk.ID()))
	diags := diag.Diagnostics{}
	diags = setResourceField(data, "storage_domain_id", string(storageDomain.ID()), diags)

	return diags
}

func deleteDisk(client ovirtclient.Client, data *schema.ResourceData) diag.Diagnostics {
	diskID := data.Get("disk_id").(string)
	storageDomainID := data.Get("storage_domain_id").(string)

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

	storageDomain, err := client.GetStorageDomain(ovirtclient.StorageDomainID(storageDomainID))
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve target storage domain.",
				Detail:   err.Error(),
			},
		}
	}

	err = client.RemoveDisk(disk.ID())
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to delete disk.",
				Detail:   err.Error(),
			},
		}
	}

	data.SetId("")
	diags := diag.Diagnostics{}
	diags = setResourceField(data, "storage_domain_id", string(storageDomain.ID()), diags)

	return diags
}

func getDisk(client ovirtclient.Client, data *schema.ResourceData) diag.Diagnostics {
	diskID := data.Get("disk_id").(string)
	storageDomainID := data.Get("storage_domain_id").(string)

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

	storageDomain, err := client.GetStorageDomain(ovirtclient.StorageDomainID(storageDomainID))
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
	diags = setResourceField(data, "storage_domain_id", string(storageDomain.ID()), diags)

	return diags
}
