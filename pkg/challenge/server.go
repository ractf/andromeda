package challenge

import (
	"sync"
	"time"
)

func StartServer(client *Client, bindIp string) *Instances {
	instances := Instances{
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

func (i *Instances) HousekeepingLoop() {
	go i.HousekeepingTick()
	ticker := time.NewTicker(time.Second * time.Duration(30))
	for {
		select {
		case <-ticker.C:
			go i.HousekeepingTick()
		}
	}
}

func (i *Instances) HousekeepingTick() {
	for _, spec := range challengeSpecList {
		instances, ok := i.challengeInstances[spec]
		if !ok {
			go i.StartInstance(spec)
			go i.StartInstance(spec)
			continue
		}

		capacity := 0
		users := 0
		count := 0
		spareInstances := 0
		for _, instance := range instances {
			if instance == nil || instance.Stopped {
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
			go i.StartInstance(spec)
		}
		if spareInstances > 2 {
			spare := &Instance{}
			found := false
			for _, instance := range instances {
				if instance == nil || instance.Stopped || len(instance.Users) > 0 {
					continue
				}
				spare = instance
				found = true
			}
			if !found {
				continue
			}

			go i.StopInstance(spare)
		}
	}
}
