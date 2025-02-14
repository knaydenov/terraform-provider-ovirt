package ovirt

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

var vmDisksMoveSchema = map[string]*schema.Schema{
	"vm_id": {
		Type:             schema.TypeString,
		Required:         true,
		ForceNew:         true,
		Description:      "Resize all disks in this VM to the specified size.",
		ValidateDiagFunc: validateUUID,
	},
	"disk_ids": {
		Type:        schema.TypeSet,
		Optional:    true,
		ForceNew:    true,
		Description: "A list of disk IDs to resize.",
		Elem: &schema.Schema{
			Type:             schema.TypeString,
			ValidateDiagFunc: validateUUID,
		},
	},
	"storage_domain_id": {
		Type:             schema.TypeString,
		Required:         true,
		Description:      "ID of the target storage domain.",
		ForceNew:         true,
		ValidateDiagFunc: validateUUID,
	},
}

func (p *provider) vmDisksMoveResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: p.vmDisksMoveCreate,
		ReadContext:   p.vmDisksMoveRead,
		DeleteContext: p.vmDisksMoveDelete,
		Schema:        vmDisksMoveSchema,
		Description:   `The ovirt_vm_disks_move resource moves all disks in an oVirt VM to the specified storage domain.`,
	}
}

func (p *provider) vmDisksMoveCreate(
	ctx context.Context,
	data *schema.ResourceData,
	_ interface{},
) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	return moveAllDisks(client, data)
}

func (p *provider) vmDisksMoveRead(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)

	targetStorageDomainID := data.Get("storage_domain_id").(string)
	currentStorageDomainID := data.Get("storage_domain_id").(string)

	vmID := ovirtclient.VMID(data.Get("vm_id").(string))

	diskAttachments, err := client.ListDiskAttachments(vmID)
	if err != nil {
		return errorToDiags(fmt.Sprintf("list disk attachments of VM %s", vmID), err)
	}
	for _, diskAttachment := range diskAttachments {
		disk, err := diskAttachment.Disk()
		if err != nil {
			return errorToDiags(fmt.Sprintf("get disk %s", diskAttachment.DiskID()), err)
		}
		targetStorageDomainID = string(disk.StorageDomainIDs()[0])

		var currentStorageDomainIds []string
		for _, storageDomainId := range disk.StorageDomainIDs() {
			currentStorageDomainIds = append(currentStorageDomainIds, string(storageDomainId))
		}
		if !inList(targetStorageDomainID, currentStorageDomainIds) {
			currentStorageDomainID = currentStorageDomainIds[0]
		}
	}

	data.SetId(string(vmID))
	diags := diag.Diagnostics{}
	diags = setResourceField(data, "storage_domain_id", currentStorageDomainID, diags)

	return diags
}

func (p *provider) vmDisksMoveDelete(_ context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	data.SetId("")
	return nil
}

func moveAllDisks(client ovirtclient.Client, data *schema.ResourceData) diag.Diagnostics {
	vmID := ovirtclient.VMID(data.Get("vm_id").(string))
	targetStorageDomainID := ovirtclient.StorageDomainID(data.Get("storage_domain_id").(string))
	diskIds := []string{}

	if diskIdsList, ok := data.GetOk("disk_ids"); ok {
		for _, diskID := range diskIdsList.(*schema.Set).List() {
			diskIds = append(diskIds, diskID.(string))
		}
	}

	diskAttachments, err := client.ListDiskAttachments(vmID)
	if err != nil {
		return errorToDiags(fmt.Sprintf("list disk attachments of VM %s", vmID), err)
	}
	var diags diag.Diagnostics
	for _, diskAttachment := range diskAttachments {
		disk, err := diskAttachment.Disk()
		if err != nil {
			return errorToDiags(fmt.Sprintf("get disk %s", diskAttachment.DiskID()), err)
		}
		if len(diskIds) > 0 && !inList(string(disk.ID()), diskIds) {
			continue
		}
		if isDiskInStorageDomain(disk, targetStorageDomainID) {
			continue
		}
		disk, err = client.MoveDiskToStorageDomain(disk.ID(), targetStorageDomainID)
		if err != nil {
			return errorToDiags(fmt.Sprintf("move disk %s", diskAttachment.DiskID()), err)
		}

		//if disk.ProvisionedSize() == desiredSize {
		//	continue
		//}
		//params := ovirtclient.UpdateDiskParams()
		//if _, err := params.WithProvisionedSize(desiredSize); err != nil {
		//	diags = append(
		//		diags,
		//		diag.Diagnostic{
		//			Severity: diag.Error,
		//			Summary:  fmt.Sprintf("Failed to set parameters for updating disk %s size.", disk.ID()),
		//			Detail:   err.Error(),
		//		},
		//	)
		//	continue
		//}
		//updateFailedDiag := diag.Diagnostic{
		//	Severity: diag.Error,
		//	Summary:  fmt.Sprintf("Failed to update disk %s size.", disk.ID()),
		//}
		//diskUpdate, err := client.StartUpdateDisk(disk.ID(), params)
		//if err != nil {
		//	updateFailedDiag.Detail = err.Error()
		//	diags = append(diags, updateFailedDiag)
		//	continue
		//}
		//_, err = diskUpdate.Wait()
		//if err != nil {
		//	updateFailedDiag.Detail = err.Error()
		//	diags = append(diags, updateFailedDiag)
		//	continue
		//}
	}

	data.SetId(string(vmID))
	if !diags.HasError() {
		diags = setResourceField(data, "storage_domain_id", targetStorageDomainID, diags)
	}
	return diags
}

func isDiskInStorageDomain(disk ovirtclient.Disk, storageDomainId ovirtclient.StorageDomainID) bool {
	for _, currentStorageDomainId := range disk.StorageDomainIDs() {
		if currentStorageDomainId == storageDomainId {
			return true
		}
	}
	return false
}
