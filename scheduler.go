package main

import(
	"log"
	"strings"
	)

const UINT_MAX = ^uint(0)

func gpuassign() string{


	var mem uint
	var gputype string
	//var id  string
	fidandmemory := getDevicesMemory()
	idandmemory := getDevicesMemory()

	podaa :=  Podinfo()

        for _,j := range(podaa){
		if j.state=="Running" && j.gputype == "memory"{
			idandmemory[j.gpuid] -= j.mem
		}else if j.state=="Running" && j.gputype != "memory"{
			 idlist := strings.Split(j.gpuid,",")
                                for _,id := range(idlist){
                                        idandmemory[id] = uint(0)
                                }

		}else if j.state=="Pending" && j.mem != uint(0){

			if j.gputype == "memory"{
				gputype = "memory"
			}else{
				gputype = "count"
			}
				mem = j.mem

		}else{
			log.Printf("pass\n")
		}

        }
	log.Printf("ASSIGN:%+v\n",idandmemory)
	//log.Printf("NeedMemory:%d\n",mem)

	best := gpuallocate(fidandmemory,idandmemory,mem,gputype)

	return best

}

func gpuallocate(data map[string]uint,data1 map[string]uint,needmem uint,gputype string)string{

	lowmem := UINT_MAX
	var bestid string
	if gputype == "memory"{
		for i,j :=range(data1){
			if j>needmem && j<lowmem {
				bestid = i
				lowmem = j
			}
		}
		return bestid
	}else{
		freegpu := 0
		for i,j := range(data1){
			if data[i] == j{
				if freegpu != 0{
					bestid += ","
					bestid += i
				}else{
					bestid += i
				}
				freegpu += 1
				if freegpu == int(needmem){
					return bestid
				}
			}
		}
	}
	return ""

}


