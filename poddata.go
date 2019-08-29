package main

import (
        "fmt"
//      "time"
        "os"
	"strconv"

        //"k8s.io/apimachinery/pkg/api/errors"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "k8s.io/client-go/kubernetes"
        "k8s.io/client-go/rest"
        //"k8s.io/apimachinery/pkg/fields"
        //"k8s.io/apimachinery/pkg/labels"
        //"k8s.io/api/core/v1"
//
        // Uncomment to load all auth plugins
        // _ "k8s.io/client-go/plugin/pkg/client/auth
        //
        // Or uncomment to load specific auth plugins
        // _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
        // _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
        // _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
        // _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

var(
        nodeName string
)

type Podfig struct{
	state string
	gputype string
	gpuid string
	mem uint
}

//nodeName = os.Getenv("NODE_NAME")


func updatenode()string{
	config, err := rest.InClusterConfig()
        if err != nil {
                panic(err.Error())
        }

        clientset, err := kubernetes.NewForConfig(config)

        if err != nil {
                panic(err.Error())
        }

	node, err := clientset.CoreV1().Nodes().Get("ndc73-0541",metav1.GetOptions{})
        if err != nil {
                fmt.Printf("Node error:%+v\n",err)
		return "fail"
        }
        a := node.ObjectMeta.Annotations

        for i,j := range(a){
                fmt.Printf("key:%s value:%s\n",i,j)
        }

        newNode := node.DeepCopy()
        newNode.ObjectMeta.Annotations["GPUInfo"] = getfakegpuid()

         _, err = clientset.CoreV1().Nodes().Update(newNode)
		if err != nil{
			fmt.Printf("Node Update Fail\n")
			return "fail"
		}
	return "Ok"

}

func updatepod(upid string)string{

	//nodeName = os.Getenv("NODE_NAME")
	config, err := rest.InClusterConfig()
        if err != nil {
                panic(err.Error())
        }

        clientset, err := kubernetes.NewForConfig(config)

        if err != nil {
                panic(err.Error())
        }
/*
	node, err := clientset.CoreV1().Nodes().Get("ndc73-0541",metav1.GetOptions{})
        if err != nil {
		fmt.Printf("Node error:%+v\n",err)
        }
        a := node.ObjectMeta.Annotations

        for i,j := range(a){
                fmt.Printf("key:%s value:%s\n",i,j)
        }

	newNode := node.DeepCopy()
	newNode.ObjectMeta.Annotations["GPUInfo"] = "test"

	 _, err = clientset.CoreV1().Nodes().Update(newNode)
                        if err != nil{
                                fmt.Printf("Node Update Fail\n")
                        }
*/

	pods, err := clientset.CoreV1().Pods("default").List(metav1.ListOptions{})

	if err != nil {
                        panic(err.Error())
                }
	for _, pod := range pods.Items {
		fmt.Printf("Update Start\n")
		fmt.Printf("Status:%s\n",pod.Status.Phase)
		if pod.Status.Phase=="Pending" && len(pod.ObjectMeta.Annotations["GPUID"])== 0{
			var total uint
			for _,me := range pod.Spec.Containers{
				if val, ok := me.Resources.Limits[resourceName]; ok {

                                        total += uint(val.Value())
                                }


			}

			newPod := pod.DeepCopy()
			if len(newPod.ObjectMeta.Annotations)== 0{
				newPod.ObjectMeta.Annotations = map[string]string{}
			}
			newPod.ObjectMeta.Annotations["GPUID"] = upid
			newPod.ObjectMeta.Annotations["GPUMEM"] = strconv.Itoa(int(total))
			//newPod.ObjectMeta.Annotations["GPUAllInfo"] = getfakegpuid()
			_, err = clientset.CoreV1().Pods("default").Update(newPod)
			if err != nil{
				fmt.Printf("Pod Update Fail\n")
			}
                        //fmt.Printf("%+v\n",pod.ObjectMeta.Annotations)
		}
	}
	//updatenode()
	return "OK"
}


func Podinfo()map[string]Podfig{
        // creates the in-cluster config
        config, err := rest.InClusterConfig()
        if err != nil {
                panic(err.Error())
        }
        // creates the clientset
        clientset, err := kubernetes.NewForConfig(config)
        if err != nil {
                panic(err.Error())
        }

        nodeName = os.Getenv("NODE_NAME")
	//podname := make(map[string]map[string]uint)
	podlist := make(map[string]Podfig)
	//podstate := make(map[string]uint)

        for ii:=0; ii<1 ;ii++ {

                //selector := fields.SelectorFromSet(fields.Set{"spec.nodeName": nodeName})//, "status.phase": "Running"})
		pods, err := clientset.CoreV1().Pods("default").List(metav1.ListOptions{
                //FieldSelector: selector.String(),
                //LabelSelector: labels.Everything().String(),
        })
                if err != nil {
                        panic(err.Error())
                }

		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

                for _, pod := range pods.Items {
                        fmt.Printf("Name: %s, Status: %s\n", pod.ObjectMeta.Name, pod.Status.Phase )

			var total uint
			flag := 0
                        for _,re := range pod.Spec.Containers{
                                if val, ok := re.Resources.Limits[resourceName]; ok {
                                        total += uint(val.Value())
					flag = 1
                                }
			}
			if flag != 0{
                                fmt.Printf("Limits:%d\n",total)
				//podstate := make(map[string]uint)
				//podstate[string(pod.Status.Phase)] = total
				var useid string
				var usetype string
				if pod.ObjectMeta.Annotations["gputype"]=="memory"{
					usetype = "memory"
				}else{
					usetype = "count"
				}

				if t := pod.ObjectMeta.Annotations["GPUID"];len(t) != 0{
					useid = t
				}else{
					useid = ""
				}
				//podname[string(pod.ObjectMeta.Name)] = podstate
				pdata := Podfig{string(pod.Status.Phase),usetype,useid,total}
				podlist[string(pod.ObjectMeta.Name)] = pdata
			}
                }
                //time.Sleep(1 * time.Second)
        }
	return podlist
}



