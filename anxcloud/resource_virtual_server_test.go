package anxcloud

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/anexia-it/go-anxcloud/pkg/vsphere"
	"github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/vm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/lithammer/shortuuid"
)

func TestAccAnxCloudVirtualServerBasic(t *testing.T) {
	resourceName := "acc_test_vm_test"
	resourcePath := "anxcloud_virtual_server." + resourceName

	vmDef := vm.Definition{
		Location:     "52b5f6b2fd3a4a7eaaedf1a7c019e9ea",
		TemplateType: "templates",
		TemplateID:   "12c28aa7-604d-47e9-83fb-5f1d1f1837b3",
		Hostname:     "acc-test-" + shortuuid.New(),
		Memory:       2048,
		CPUs:         2,
		Disk:         50,
		Network: []vm.Network{
			{
				VLAN:    "02f39d20ca0f4adfb5032f88dbc26c39",
				NICType: "vmxnet3",
			},
		},
		DNS1:     "8.8.8.8",
		Password: "flatcar#1234$%%",
	}

	vmDefUpdate := vmDef
	vmDefUpdate.CPUs = 4
	vmDefUpdate.Disk = 80
	vmDefUpdate.Memory = 4096
	vmDefUpdate.Network = append(vmDefUpdate.Network, vm.Network{
		VLAN:    "ff70791b398e4ab29786dd34f211694c",
		NICType: "vmxnet3",
	})

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAnxCloudVirtualServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigAnxCloudVirtualServer(resourceName, &vmDef),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnxCloudVirtualServerExists(resourcePath, &vmDef),
					resource.TestCheckResourceAttr(resourcePath, "location_id", vmDef.Location),
					resource.TestCheckResourceAttr(resourcePath, "template_id", vmDef.TemplateID),
					resource.TestCheckResourceAttr(resourcePath, "cpus", strconv.Itoa(vmDef.CPUs)),
					resource.TestCheckResourceAttr(resourcePath, "memory", strconv.Itoa(vmDef.Memory)),
					resource.TestCheckResourceAttr(resourcePath, "disk", strconv.Itoa(vmDef.Disk)),
				),
			},
			{
				Config: testAccConfigAnxCloudVirtualServer(resourceName, &vmDefUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnxCloudVirtualServerExists(resourcePath, &vmDefUpdate),
					resource.TestCheckResourceAttr(resourcePath, "location_id", vmDefUpdate.Location),
					resource.TestCheckResourceAttr(resourcePath, "template_id", vmDefUpdate.TemplateID),
					resource.TestCheckResourceAttr(resourcePath, "cpus", strconv.Itoa(vmDefUpdate.CPUs)),
					resource.TestCheckResourceAttr(resourcePath, "memory", strconv.Itoa(vmDefUpdate.Memory)),
					resource.TestCheckResourceAttr(resourcePath, "disk", strconv.Itoa(vmDefUpdate.Disk)),
				),
			},
		},
	})
}

func testAccCheckAnxCloudVirtualServerDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(client.Client)
	v := vsphere.NewAPI(c)
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "anxcloud_virtual_server" {
			continue
		}

		if rs.Primary.ID == "" {
			return nil
		}

		info, err := v.Info().Get(ctx, rs.Primary.ID)
		if err != nil {
			if err := handleNotFoundError(err); err != nil {
				return err
			}
			return nil
		}
		if info.Identifier != "" {
			return fmt.Errorf("virtual machine '%s' exists", info.Identifier)
		}
	}

	return nil
}

func testAccConfigAnxCloudVirtualServer(resourceName string, def *vm.Definition) string {
	return fmt.Sprintf(`
	resource "anxcloud_virtual_server" "%s" {
		location_id   = "%s"
		template_id   = "%s"
		template_type = "%s"
		hostname      = "%s"
		cpus          = %d
		memory        = %d
		disk          = %d
		password      = "%s"

		// generated network string
		%s

		force_restart_if_needed = true
		critical_operation_confirmed = true
	}
	`, resourceName, def.Location, def.TemplateID, def.TemplateType, def.Hostname, def.CPUs, def.Memory, def.Disk,
		def.Password, generateNetworkSubResourceString(def.Network))
}

func testAccCheckAnxCloudVirtualServerExists(n string, def *vm.Definition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		c := testAccProvider.Meta().(client.Client)
		v := vsphere.NewAPI(c)
		ctx := context.Background()

		if !ok {
			return fmt.Errorf("virtual server not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("virtual server id not set")
		}

		info, err := v.Info().Get(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if info.Status != vmPoweredOn {
			return fmt.Errorf("virtual machine found but it is not in the expected state '%s': '%s'", vmPoweredOn, info.Status)
		}

		if info.CPU != def.CPUs {
			return fmt.Errorf("virtual machine cpu does not match, got %d - expected %d", info.CPU, def.CPUs)
		}
		if info.RAM != def.Memory {
			return fmt.Errorf("virtual machine memory does not match, got %d - expected %d", info.RAM, def.Memory)
		}

		if len(info.DiskInfo) != 1 {
			return fmt.Errorf("unspported number of attached disks, got %d - expected 1", len(info.DiskInfo))
		}
		if infoDiskGB := info.DiskInfo[0].DiskGB; infoDiskGB != def.Disk {
			return fmt.Errorf("virtual machine disk size does not match, got %d - expected %d", infoDiskGB, def.Disk)
		}

		if len(info.Network) != len(def.Network) {
			return fmt.Errorf("virtual machine networks number do not match, got %d - expected %d", len(info.Network), len(info.Network))
		}
		for i, n := range def.Network {
			if n.VLAN != info.Network[i].VLAN {
				return fmt.Errorf("virtual machine network[%d].vlan do not match, got %s - expected %s", i, info.Network[i].VLAN, n.VLAN)
			}
		}

		return nil
	}
}

func generateNetworkSubResourceString(networks []vm.Network) string {
	var output string
	template := "\nnetwork {\n\tvlan_id = \"%s\"\n\tnic_type = \"%s\"\n}\n"

	for _, n := range networks {
		output += fmt.Sprintf(template, n.VLAN, n.NICType)
	}

	return output
}
