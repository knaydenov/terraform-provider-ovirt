package ovirt

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func (p *provider) diskDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: p.diskDataSourceRead,
		Schema:      diskBaseSchema,
		Description: `Returns the disk.`,
	}
}

func (p *provider) diskDataSourceRead(
	ctx context.Context,
	data *schema.ResourceData,
	_ interface{},
) diag.Diagnostics {
	client := p.client.WithContext(ctx)

	disk, err := client.GetDisk(ovirtclient.DiskID(data.Get("vm_id").(string)))
	if err != nil {
		return errorToDiags("getting disk", err)
	}

	data.SetId(string(disk.ID()))
	err = data.Set("storage_domain_id", disk.StorageDomainIDs()[0])
	if err != nil {
		return errorToDiags("getting storage_domain_id", err)
	}
	err = data.Set("format", string(disk.Format()))
	if err != nil {
		return errorToDiags("getting format", err)
	}
	err = data.Set("alias", disk.Alias())
	if err != nil {
		return errorToDiags("getting alias", err)
	}
	err = data.Set("sparse", disk.Sparse())
	if err != nil {
		return errorToDiags("getting sparse", err)
	}
	err = data.Set("sparse", disk.Sparse())
	if err != nil {
		return errorToDiags("getting sparse", err)
	}
	err = data.Set("total_size", disk.TotalSize())
	if err != nil {
		return errorToDiags("getting total_size", err)
	}
	err = data.Set("status", string(disk.Status()))
	if err != nil {
		return errorToDiags("getting status", err)
	}

	return nil
}
