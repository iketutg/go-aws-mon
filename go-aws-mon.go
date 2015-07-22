package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"log"
	"os"
	"strings"
)

func main() {
	isMemUtil := flag.Bool("mem-util", true, "Memory Utilization(percent)")
	isMemUsed := flag.Bool("mem-used", false, "Memory Used(bytes)")
	isMemAvail := flag.Bool("mem-avail", false, "Memory Available(bytes)")
	isSwapUtil := flag.Bool("swap-util", false, "Swap Utilization(percent)")
	isSwapUsed := flag.Bool("swap-used", false, "Swap Used(bytes)")
	isDiskSpaceUtil := flag.Bool("disk-space-util", true, "Disk Space Utilization(percent)")
	isDiskSpaceUsed := flag.Bool("disk-space-used", false, "Disk Space Used(bytes)")
	isDiskSpaceAvail := flag.Bool("disk-space-avail", false, "Disk Space Available(bytes)")
	isDiskInodeUtil := flag.Bool("disk-inode-util", false, "Disk Inode Utilization(percent)")

	ns := flag.String("namespace", "Linux/System", "CloudWatch metric namespace (required)(It is always EC2)")
	diskPaths := flag.String("disk-path", "/", "Disk Path")

	flag.Parse()

	metadata, err := getInstanceMetadata()

	if err != nil {
		log.Fatal("Can't get InstanceData, please confirm we are running on a AWS EC2 instance: ", err)
		os.Exit(1)
	}

	for k, v := range metadata {
		fmt.Println(k, " : ", v)
	}

	memUtil, memUsed, memAvail, swapUtil, swapUsed, err := memoryUsage()

	var metricData []*cloudwatch.MetricDatum

	dims := getDimensions(metadata)
	if *isMemUtil {
		metricData, err = addMetric("MemoryUtilization", "Percent", memUtil, dims, metricData)
		if err != nil {
			log.Fatal("Can't add memory usage metric: ", err)
		}
	}

	if *isMemUsed {
		metricData, err = addMetric("MemoryUsed", "Bytes", memUsed, dims, metricData)
		if err != nil {
			log.Fatal("Can't add memory used metric: ", err)
		}
	}
	if *isMemAvail {
		metricData, err = addMetric("MemoryAvail", "Bytes", memAvail, dims, metricData)
		if err != nil {
			log.Fatal("Can't add memory available metric: ", err)
		}
	}
	if *isSwapUsed {
		metricData, err = addMetric("SwapUsed", "Bytes", swapUsed, dims, metricData)
		if err != nil {
			log.Fatal("Can't add swap used metric: ", err)
		}
	}
	if *isSwapUtil {
		metricData, err = addMetric("SwapUtil", "Percent", swapUtil, dims, metricData)
		if err != nil {
			log.Fatal("Can't add swap usage metric: ", err)
		}
	}

	paths := strings.Split(*diskPaths, ",")

	for _, val := range paths {
		diskspaceUtil, diskspaceUsed, diskspaceAvail, diskinodesUtil, err := DiskSpace(val)
		if err != nil {
			log.Fatal("Can't get DiskSpace %s", err)
		}
		metadata["FileSystem"] = val
		dims := getDimensions(metadata)
		if *isDiskSpaceUtil {
			metricData, err = addMetric("DiskUtilization", "Percent", diskspaceUtil, dims, metricData)
			if err != nil {
				log.Fatal("Can't add Disk Utilization metric: ", err)
			}
		}
		if *isDiskSpaceUsed {
			metricData, err = addMetric("DiskUsed", "Bytes", float64(diskspaceUsed), dims, metricData)
			if err != nil {
				log.Fatal("Can't add Disk Used metric: ", err)
			}
		}
		if *isDiskSpaceAvail {
			metricData, err = addMetric("DiskAvail", "Bytes", float64(diskspaceAvail), dims, metricData)
			if err != nil {
				log.Fatal("Can't add Disk Available metric: ", err)
			}
		}
		if *isDiskInodeUtil {
			metricData, err = addMetric("DiskInodesUtilization", "Percent", diskinodesUtil, dims, metricData)
			if err != nil {
				log.Fatal("Can't add Disk Inodes Utilization metric: ", err)
			}
		}
	}

	for _, mData := range metricData {
		fmt.Printf("%+v\n", *mData)
	}

	err = putMetric(metricData, *ns, metadata["region"])
	if err != nil {
		log.Fatal("Can't put CloudWatch Metric")
	}
}