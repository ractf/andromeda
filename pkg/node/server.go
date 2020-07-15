package node

import (
	"sync"
	"time"
)

func StartServer(client *Client, bindIp string) *Node {
	instances := Node{
		Client:             client,
		challengeInstances: make(map[Spec][]*Instance),
		userInstances:      make(map[string]*Instance),
		userAvoids:         make(map[string]*Instance),
		mutex:              &sync.Mutex{},
		bindIp:             bindIp,
	}
	go instances.HousekeepingLoop()
	return &instances
}

func (n *Node) HousekeepingLoop() {
	go n.HousekeepingTick()
	ticker := time.NewTicker(time.Second * time.Duration(30))
	for {
		select {
		case <-ticker.C:
			go n.HousekeepingTick()
		}
	}
}

func (n *Node) HousekeepingTick() {
	for _, spec := range challengeSpecList {
		instances, ok := n.challengeInstances[spec]
		if !ok {
			go n.StartInstance(spec)
			go n.StartInstance(spec)
			continue
		}

		capacity := 0
		users := 0
		count := 0
		spareInstances := 0
		for _, instance := range instances {
			if instance.Stopped {
				continue
			}
			capacity += spec.UserLimit
			users += len(instance.Users)
			count++
			if len(instance.Users) == 0 {
				spareInstances++
			}
		}

		if spareInstances == 0 {
			go n.StartInstance(spec)
		}
		if spareInstances > 2 {
			spare := &Instance{}
			found := false
			for _, instance := range instances {
				if instance.Stopped || len(instance.Users) > 0 {
					continue
				}
				spare = instance
				found = true
			}
			if !found {
				continue
			}

			go n.StopInstance(spare)
		}
	}
}
