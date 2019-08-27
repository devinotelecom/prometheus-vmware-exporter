package controllers

import (
	"context"
	"strings"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/govmomi/find"
)

const namespace = "vmware"

var (
	prometheusHostPowerState = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "host",
		Name:      "power_state",
		Help:      "poweredOn 1, poweredOff 2, standBy 3, other 0",
	}, []string{"host_name"})
	prometheusHostBoot = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "host",
		Name:      "boot_timestamp_seconds",
		Help:      "Uptime host",
	}, []string{"host_name"})
	prometheusTotalCpu = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "host",
		Name:      "cpu_max",
		Help:      "CPU total",
	}, []string{"host_name"})
	prometheusUsageCpu = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "host",
		Name:      "cpu_usage",
		Help:      "CPU Usage",
	}, []string{"host_name"})
	prometheusTotalMem = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "host",
		Name:      "memory_max",
		Help:      "Memory max",
	}, []string{"host_name"})
	prometheusUsageMem = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "host",
		Name:      "memory_usage",
		Help:      "Memory Usage",
	}, []string{"host_name"})
	prometheusDiskOk = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "host",
		Name:      "disk_ok",
		Help:      "Disk is working normally",
	}, []string{"host_name", "device"})
	prometheusStorageStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "host",
		Name:      "storage_status",
		Help:      "green 0, yellow 1, red 2",
	}, []string{"host_name", "device"})
	prometheusTotalDs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "datastore",
		Name:      "capacity_size",
		Help:      "Datastore total",
	}, []string{"ds_name", "host_name"})
	prometheusUsageDs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "datastore",
		Name:      "freespace_size",
		Help:      "Datastore free",
	}, []string{"ds_name", "host_name"})
	prometheusVmBoot = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "boot_timestamp_seconds",
		Help:      "VMWare VM boot time in seconds",
	}, []string{"vm_name", "host_name"})
	prometheusVmCpuAval = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "cpu_avaleblemhz",
		Help:      "VMWare VM usage CPU",
	}, []string{"vm_name", "host_name"})
	prometheusVmCpuUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "cpu_usagemhz",
		Help:      "VMWare VM usage CPU",
	}, []string{"vm_name", "host_name"})
	prometheusVmNumCpu = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "num_cpu",
		Help:      "Available number of cores",
	}, []string{"vm_name", "host_name"})
	prometheusVmMemAval = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "mem_avaleble",
		Help:      "Available memory",
	}, []string{"vm_name", "host_name"})
	prometheusVmMemUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "mem_usage",
		Help:      "Usage memory",
	}, []string{"vm_name", "host_name"})
	prometheusVmNetRec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "net_rec",
		Help:      "Usage memory",
	}, []string{"vm_name", "host_name"})
)

func totalCpu(hs mo.HostSystem) float64 {
	totalCPU := int64(hs.Summary.Hardware.CpuMhz) * int64(hs.Summary.Hardware.NumCpuCores)
	return float64(totalCPU)
}

func convertTime(vm mo.VirtualMachine) float64 {
	if vm.Summary.Runtime.BootTime == nil {
		return 0
	}
	return float64(vm.Summary.Runtime.BootTime.Unix())
}

func powerState(s types.HostSystemPowerState) float64 {
	if s == "poweredOn" {
		return 1
	}
	if s == "poweredOff" {
		return 2
	}
	if s == "standBy" {
		return 3
	}
	return 0
}

func RegistredMetrics() {
	prometheus.MustRegister(
		prometheusHostPowerState,
		prometheusHostBoot,
		prometheusTotalCpu,
		prometheusUsageCpu,
		prometheusTotalMem,
		prometheusUsageMem,
		prometheusDiskOk,
		prometheusStorageStatus,
		prometheusTotalDs,
		prometheusUsageDs,
		prometheusVmBoot,
		prometheusVmCpuAval,
		prometheusVmNumCpu,
		prometheusVmMemAval,
		prometheusVmMemUsage,
		prometheusVmCpuUsage,
		prometheusVmNetRec)
}

func NewVmwareHostMetrics(host string, username string, password string, logger *log.Logger) {
	ctx := context.Background()
	c, err := NewClient(ctx, host, username, password)
	if err != nil {
		logger.Fatal(err)
	}
	defer c.Logout(ctx)
	m := view.NewManager(c.Client)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		logger.Fatal(err)
	}
	defer v.Destroy(ctx)
	var hss []mo.HostSystem
	err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"summary"}, &hss)
	if err != nil {
		logger.Fatal(err)
	}
	for _, hs := range hss {
		for _, ssi := range hs.Summary.Runtime.HealthSystemRuntime.HardwareStatusInfo.StorageStatusInfo {
			name := ssi.Name
			if strings.Count(name, ":") == 4 && strings.Contains(name, "Disk") {
				// Remove disk status from name to make the name more stable
				name = name[:strings.LastIndex(name, ":")]
			}
			name = strings.TrimSpace(name)
			status := -1
			switch ssi.Status.GetElementDescription().Key {
			case "Green":
				status = 0
			case "Yellow":
				status = 1
			case "Red":
				status = 2
			}
			if (status != -1) {
				prometheusStorageStatus.WithLabelValues(host, name).Set(float64(status))
			}
		}
		prometheusHostPowerState.WithLabelValues(host).Set(powerState(hs.Summary.Runtime.PowerState))
		prometheusHostBoot.WithLabelValues(host).Set(float64(hs.Summary.Runtime.BootTime.Unix()))
		prometheusTotalCpu.WithLabelValues(host).Set(totalCpu(hs))
		prometheusUsageCpu.WithLabelValues(host).Set(float64(hs.Summary.QuickStats.OverallCpuUsage))
		prometheusTotalMem.WithLabelValues(host).Set(float64(hs.Summary.Hardware.MemorySize))
		prometheusUsageMem.WithLabelValues(host).Set(float64(hs.Summary.QuickStats.OverallMemoryUsage) * 1024 * 1024)

	}
	finder := find.NewFinder(c.Client)
	hs, err := finder.DefaultHostSystem(ctx)
	if err != nil {
		logger.Fatal(err)
	}
	ss, err := hs.ConfigManager().StorageSystem(ctx)
	if err != nil {
		logger.Fatal(err)
	}
	var hostss mo.HostStorageSystem
	err = ss.Properties(ctx, ss.Reference(), nil, &hostss)
	if err != nil {
		logger.Fatal(err)
	}
	for _, e := range hostss.StorageDeviceInfo.ScsiLun {
		lun := e.GetScsiLun()
		ok := 1.0
		for _, s := range lun.OperationalState {
			if s != "ok" {
				ok = 0
				break
			}
		}
		prometheusDiskOk.WithLabelValues(host, lun.DeviceName).Set(ok)
	}
}

func NewVmwareDsMetrics(host string, username string, password string, logger *log.Logger) {
	ctx := context.Background()
	c, err := NewClient(ctx, host, username, password)
	if err != nil {
		logger.Fatal(err)
	}
	defer c.Logout(ctx)
	m := view.NewManager(c.Client)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Datastore"}, true)
	if err != nil {
		logger.Fatal(err)
	}
	defer v.Destroy(ctx)
	var dss []mo.Datastore
	err = v.Retrieve(ctx, []string{"Datastore"}, []string{"summary"}, &dss)
	if err != nil {
		logger.Fatal(err)
	}
	for _, ds := range dss {
		dsname := ds.Summary.Name
		prometheusTotalDs.WithLabelValues(dsname, host).Set(float64(ds.Summary.Capacity))
		prometheusUsageDs.WithLabelValues(dsname, host).Set(float64(ds.Summary.FreeSpace))
	}
}

func NewVmwareVmMetrics(host string, username string, password string, logger *log.Logger) {
	ctx := context.Background()
	c, err := NewClient(ctx, host, username, password)
	if err != nil {
		logger.Fatal(err)
	}
	defer c.Logout(ctx)
	m := view.NewManager(c.Client)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		logger.Fatal(err)
	}
	defer v.Destroy(ctx)
	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary"}, &vms)
	if err != nil {
		logger.Fatal(err)
	}
	for _, vm := range vms {
		vmname := vm.Summary.Config.Name
		prometheusVmBoot.WithLabelValues(vmname, host).Set(convertTime(vm))
		prometheusVmCpuAval.WithLabelValues(vmname, host).Set(float64(vm.Summary.Runtime.MaxCpuUsage) * 1000 * 1000)
		prometheusVmCpuUsage.WithLabelValues(vmname, host).Set(float64(vm.Summary.QuickStats.OverallCpuUsage) * 1000 * 1000)
		prometheusVmNumCpu.WithLabelValues(vmname, host).Set(float64(vm.Summary.Config.NumCpu))
		prometheusVmMemAval.WithLabelValues(vmname, host).Set(float64(vm.Summary.Config.MemorySizeMB))
		prometheusVmMemUsage.WithLabelValues(vmname, host).Set(float64(vm.Summary.QuickStats.GuestMemoryUsage) * 1024 * 1024)
	}
}
