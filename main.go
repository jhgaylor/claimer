package main

import (
	"github.com/hashicorp/consul/api"
	"fmt"
	"os"
)

func main () {
	hostname, _ := os.Hostname()
	consulClient, err := api.NewClient(&api.Config{Address: "127.0.0.1:8500"})

	opts := &api.LockOptions{
	  Key:        "master_acquisition",
	  Value:      []byte(fmt.Sprintf("set by %s", hostname)),
	  SessionTTL: "30s",
	  // TODO: use health checks to release lock
	}
	lock, err := consulClient.LockOpts(opts)

	// TODO: error handling
	stopCh := make(chan struct{})
	lock.Lock(stopCh)
	
	masterTaken := false
	masterServiceName := "k8s-master"
	masterServiceResults, _, err := consulClient.Catalog().Service(masterServiceName, "", nil)
	if (err != nil) {
		fmt.Printf("dude! service retreival failed. %+v\n", err)
	}
  for _, serviceResult := range masterServiceResults {
  	
  	if hostname == serviceResult.Node {
  		masterTaken = true
  	}
  }
  fmt.Printf("Master taken: %s", masterTaken)

  currentServiceName := "k8s-node"
	if ! masterTaken {
	    currentServiceName = masterServiceName
	}
	serviceRegistration := &api.AgentServiceRegistration{
	    Name: currentServiceName,
	}
	err = consulClient.Agent().ServiceRegister(serviceRegistration)
	if (err != nil) {
		fmt.Printf("Error registering consul service: %s", err)
	}
	lock.Unlock()
}
