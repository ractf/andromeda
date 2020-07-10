package challenge

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
)

type Instances struct {
	Client             *Client
	challengeInstances map[Spec][]*Instance
	userInstances      map[string]*Instance
	userAvoids         map[string]*Instance
	mutex              *sync.Mutex
	bindIp             string
}

type Instance struct {
	Port      string
	Challenge Spec
	Users     []string
	Container string
	Stopped   bool
	avoiding  *[]string
}

func contains(haystack *[]string, needle string) (bool, int) {
	for i, element := range *haystack {
		if element == needle {
			return true, i
		}
	}
	return false, 0
}

func (i *Instances) GetChallengeByName(name string) Spec {
	return challengeSpecs[name]
}

func remove(slice []string, s int) []string {
	return append(slice[:s], slice[s+1:]...)
}

func (i *Instances) GetInstanceForUser(user string, challenge Spec) (*Instance, error) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if instance, ok := i.userInstances[user]; ok && instance.Challenge == challenge {
		return instance, nil
	}

	instances, ok := i.challengeInstances[challenge]
	if !ok || len(instances) == 0 {
		return &Instance{}, errors.New("no instances of challenge")
	}
	shuffledInstances := make([]*Instance, len(instances))
	copy(shuffledInstances, instances)

	rand.Shuffle(len(shuffledInstances), func(i, j int) {
		shuffledInstances[i], shuffledInstances[j] = shuffledInstances[j], shuffledInstances[i]
	})

	for _, instance := range shuffledInstances {
		if !instance.Stopped && len(instance.Users) < instance.Challenge.UserLimit && len(*instance.avoiding) < 50 {
			avoiding, _ := contains(instance.avoiding, user)
			if avoiding {
				continue
			}
			instance.Users = append(instance.Users, user)
			i.userInstances[user] = instance
			return instance, nil
		}
	}

	for _, instance := range shuffledInstances {
		if !instance.Stopped && len(instance.Users) < instance.Challenge.UserLimit {
			if avoid, ok := i.userAvoids[user]; ok && avoid == instance {
				continue
			}

			instance.Users = append(instance.Users, user)
			i.userInstances[user] = instance
			avoiding, index := contains(instance.avoiding, user)
			if avoiding {
				x := remove(*instance.avoiding, index)
				instance.avoiding = &x
			}
			return instance, nil
		}
	}

	return &Instance{}, errors.New("could not find instance")
}

func (i *Instances) GetCurrentUserInstance(user string) *Instance {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if instance, ok := i.userInstances[user]; ok {
		return instance
	}
	x := make([]string, 0)
	return &Instance{avoiding: &x}
}

func (i *Instances) StartInstance(challenge Spec) {
	fmt.Println("Starting an instance of", challenge.Name)
	instance, err := i.Client.StartContainer(challenge, i.bindIp)
	if err != nil {
		fmt.Println(err)
		return
	}

	i.mutex.Lock()
	instances, ok := i.challengeInstances[challenge]
	if !ok {
		instances = make([]*Instance, 0)
	}
	i.challengeInstances[challenge] = append(instances, &instance)
	i.mutex.Unlock()
}

func (i *Instances) AvoidInstance(user string, instance *Instance) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	avoiding := append(*instance.avoiding, user)
	instance.avoiding = &avoiding
	i.userAvoids[user] = instance
}

func (i *Instances) Disconnect(user string) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	instance := i.GetCurrentUserInstance(user)
	if instance.Users == nil {
		return
	}

	index := -1
	for j, userElement := range instance.Users {
		if user == userElement {
			index = j
			break
		}
	}
	if index != -1 {
		instance.Users = append(instance.Users[:index], instance.Users[index+1:]...)
	}

	delete(i.userInstances, user)
}

func (i *Instances) StopInstance(instance *Instance) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	_ = i.Client.StopContainer(instance.Container)
	instance.Stopped = true
}
