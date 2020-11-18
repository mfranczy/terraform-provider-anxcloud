package anxcloud

import (
	"testing"

	"github.com/anexia-it/go-anxcloud/pkg/vsphere/info"
	"github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/vm"
	"github.com/google/go-cmp/cmp"
)

// expanders tests

func TestExpanderVirtualServerNetworks(t *testing.T) {
	cases := []struct {
		Input          []interface{}
		ExpectedOutput []vm.Network
	}{
		{
			[]interface{}{
				map[string]interface{}{
					"vlan_id":  "38f8561acfe34qc49c336d2af31a5cc3",
					"nic_type": "vmxnet3",
					"ips": []interface{}{
						"identifier1",
						"identifier2",
						"10.11.12.13",
						"1.0.0.1",
					},
				},
			},
			[]vm.Network{
				{
					VLAN:    "38f8561acfe34qc49c336d2af31a5cc3",
					NICType: "vmxnet3",
					IPs: []string{
						"identifier1",
						"identifier2",
						"10.11.12.13",
						"1.0.0.1",
					},
				},
			},
		},
	}

	for _, tc := range cases {
		output := expandVirtualServerNetworks(tc.Input)
		if diff := cmp.Diff(tc.ExpectedOutput, output); diff != "" {
			t.Fatalf("Unexpected output from expander: mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestExpanderVirtualServerDNS(t *testing.T) {
	cases := []struct {
		Input          []interface{}
		ExpectedOutput [maxDNSEntries]string
	}{
		{
			[]interface{}{
				"1.1.1.1",
				"2.2.2.2",
				"3.3.3.3",
				"4.4.4.4",
			},
			[maxDNSEntries]string{
				"1.1.1.1",
				"2.2.2.2",
				"3.3.3.3",
				"4.4.4.4",
			},
		},
		{
			[]interface{}{
				"1.1.1.1",
				"2.2.2.2",
			},
			[maxDNSEntries]string{
				"1.1.1.1",
				"2.2.2.2",
				"",
				"",
			},
		},
		{
			[]interface{}{},
			[maxDNSEntries]string{
				"",
				"",
				"",
				"",
			},
		},
	}

	for _, tc := range cases {
		output := expandVirtualServerDNS(tc.Input)
		if diff := cmp.Diff(tc.ExpectedOutput, output); diff != "" {
			t.Fatalf("Unexpected output from expander: mismatch (-want +got):\n%s", diff)
		}
	}
}

// flatteners tests

func TestFlattenVirtualServerInfo(t *testing.T) {
	cases := []struct {
		Input          info.Info
		ExpectedOutput []interface{}
	}{
		{
			// TODO: fix disk and networks structs in go-client
			info.Info{
				Name:            "12345-test",
				CustomName:      "test-vm",
				Identifier:      "1111111111111111111111",
				GuestOS:         "Ubuntu Linux (64-bit)",
				LocationCode:    "ANX04",
				LocationCountry: "AT",
				LocationName:    "ANX04 - AT, Vienna, Datasix",
				Status:          "poweredOn",
				RAM:             4096,
				CPU:             4,
				Cores:           4,
				Network: []struct {
					NIC        int      `json:"nic"`
					ID         int      `json:"id"`
					VLAN       string   `json:"vlan"`
					MACAddress string   `json:"mac_address"`
					IPv4       []string `json:"ips_v4"`
					IPv6       []string `json:"ips_v6"`
				}{
					{
						NIC:        3,
						ID:         4000,
						VLAN:       "111111111111111111111",
						MACAddress: "00:50:56:bb:c0:81",
						IPv4:       []string{"1.1.1.1"},
						IPv6:       []string{"2001:db8::8a2e:370:7334"},
					},
				},
				Disks: 1,
				DiskInfo: []struct {
					DiskType     string `json:"disk_type"`
					StorageType  string `json:"storage_type"`
					BusType      string `json:"bus_type"`
					BusTypeLabel string `json:"bus_type_label"`
					DiskGB       int    `json:"disk_gb"`
					DiskID       int    `json:"disk_id"`
					IOPS         int    `json:"iops"`
					Latency      int    `json:"latence"`
				}{
					{
						DiskType:     "HPC5",
						StorageType:  "SSD",
						BusType:      "SCSI",
						BusTypeLabel: "SCSI(0:0) Hard disk 1",
						DiskGB:       90,
						DiskID:       2000,
						IOPS:         150000,
						Latency:      7,
					},
				},
				VersionTools:     "guestToolsUnmanaged",
				GuestToolsStatus: "Active",
			},
			[]interface{}{
				map[string]interface{}{
					"name":             "12345-test",
					"custom_name":      "test-vm",
					"guest_os":         "Ubuntu Linux (64-bit)",
					"location_code":    "ANX04",
					"location_country": "AT",
					"location_name":    "ANX04 - AT, Vienna, Datasix",
					"status":           "poweredOn",
					"network": []map[string]interface{}{
						{
							"nic":         3,
							"id":          4000,
							"vlan":        "111111111111111111111",
							"mac_address": "00:50:56:bb:c0:81",
							"ip_v4":       []string{"1.1.1.1"},
							"ip_v6":       []string{"2001:db8::8a2e:370:7334"},
						},
					},
					"ram":          4096,
					"cpu":          4,
					"cores":        4,
					"disks_number": 1,
					"disks_info": []map[string]interface{}{
						{
							"disk_type":      "HPC5",
							"storage_type":   "SSD",
							"bus_type":       "SCSI",
							"bus_type_label": "SCSI(0:0) Hard disk 1",
							"disk_gb":        90,
							"disk_id":        2000,
							"iops":           150000,
							"latency":        7,
						},
					},
					"version_tools":      "guestToolsUnmanaged",
					"guest_tools_status": "Active",
				},
			},
		},
	}

	for _, tc := range cases {
		output := flattenVirtualServerInfo(&tc.Input)
		if diff := cmp.Diff(tc.ExpectedOutput, output); diff != "" {
			t.Fatalf("Unexpected output from expander: mismatch (-want +got):\n%s", diff)
		}
	}
}