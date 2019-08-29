// Copyright (c) 2017, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"log"
	"strings"
	"fmt"
        "strconv"
	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"

	//"k8s-device-plugin/gpu-monitoring-tools/bindings/go/nvml"

	"golang.org/x/net/context"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

func check(err error) {
	if err != nil {
		log.Panicln("Fatal:", err)
	}
}

func combineIDcount(realID string, mCounter uint) string {
	return fmt.Sprintf("%s-_-%d", realID, mCounter)
}

func getDevicesMemory() map[string]uint {
	n, err := nvml.GetDeviceCount()
        check(err)

	m :=make(map[string]uint)

        for i := uint(0); i < n; i++ {
                d, err := nvml.NewDevice(i)
                check(err)

                log.Println("Start-GPUID:",d.UUID,"Memory:",uint(*d.Memory))
		m[d.UUID] = uint(*d.Memory)
                //log.Println("GPUid:%s Memory:%d",d.UUID,uint(d.Memory))
        }

        return m

}

func getfakegpuid() string{

	f := getDevicesMemory()
	var fakeid string

	for i,j := range(f){
		//r := fmt.Sprintf("%s-%s",i,strconv.Itoa(int(j)))
		//g := strings.Join([]string{i,strconv.Itoa(int(j))},"-")
		if len(fakeid)==0{
			fakeid = i + ":" + strconv.Itoa(int(j))

		}else{
			fakeid = fakeid + "," + i + ":" + strconv.Itoa(int(j))
			//fakeid = strings.Join([]string{fakeid,g},",")
		}

	}
	return fakeid

}



func getDevices() []*pluginapi.Device {
	n, err := nvml.GetDeviceCount()
	check(err)



	var devs []*pluginapi.Device
	for i := uint(0); i < n; i++ {
		d, err := nvml.NewDevice(i)
		check(err)

		log.Println("GPUID:",d.UUID,"Power:",uint(*d.Power),"Memory:",uint(*d.Memory))

		/*
		devs = append(devs, &pluginapi.Device{
			ID:     d.UUID,
			Health: pluginapi.Healthy,
		})
		//log.Printf("%+v\n",d)
		*/
		
		for j := uint(0); j < uint(*d.Memory); j++{


			newID := combineIDcount(d.UUID,j)
			devs = append(devs, &pluginapi.Device{
				ID:     newID,
				Health: pluginapi.Healthy,
			})

		}
	}

	return devs
}

func deviceExists(devs []*pluginapi.Device, id string) bool {
	for _, d := range devs {
		//log.Printf("ID:%d - id:%d",d.ID,id)
		if d.ID == id {
			return true
		}
	}
	return false
}

func splitDeviceID(rawID string) string {
	return strings.Split(rawID, "-_-")[0]
}

func watchXIDs(ctx context.Context, devs []*pluginapi.Device, xids chan<- *pluginapi.Device) {
	eventSet := nvml.NewEventSet()
	defer nvml.DeleteEventSet(eventSet)

 
	for _, d := range devs {
		realID := splitDeviceID(d.ID)
		err := nvml.RegisterEventForDevice(eventSet, nvml.XidCriticalError, realID)
		if err != nil && strings.HasSuffix(err.Error(), "Not Supported") {
			log.Printf("Warning: %s is too old to support healthchecking: %s. Marking it unhealthy.", d.ID, err)

			xids <- d
			continue
		}

		if err != nil {
			log.Panicln("Fatalaaaa:", err)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		e, err := nvml.WaitForEvent(eventSet, 5000)
		if err != nil && e.Etype != nvml.XidCriticalError {
			continue
		}

		// FIXME: formalize the full list and document it.
		// http://docs.nvidia.com/deploy/xid-errors/index.html#topic_4
		// Application errors: the GPU should still be healthy
		if e.Edata == 31 || e.Edata == 43 || e.Edata == 45 {
			continue
		}

		if e.UUID == nil || len(*e.UUID) == 0 {
			// All devices are unhealthy
			for _, d := range devs {
				xids <- d
			}
			continue
		}

		for _, d := range devs {
			if d.ID == *e.UUID {
				xids <- d
			}
		}
	}
}
