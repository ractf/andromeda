package node

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
)

type Node struct {
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

func (n *Node) GetChallengeByName(name string) Spec {
	return challengeSpecs[name]
}

func remove(slice []string, s int) []string {
	return append(slice[:s], slice[s+1:]...)
}

func (n *Node) GetInstanceForUser(user string, challenge Spec) (*Instance, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if instance, ok := n.userInstances[user]; ok && instance.Challenge == challenge {
		return instance, nil
	}

	instances, ok := n.challengeInstances[challenge]
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
			n.userInstances[user] = instance
			return instance, nil
		}
	}

	for _, instance := range shuffledInstances {
		if !instance.Stopped && len(instance.Users) < instance.Challenge.UserLimit {
			if avoid, ok := n.userAvoids[user]; ok && avoid == instance {
				continue
			}

			instance.Users = append(instance.Users, user)
			n.userInstances[user] = instance
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

func (n *Node) GetCurrentUserInstance(user string) *Instance {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if instance, ok := n.userInstances[user]; ok {
		return instance
	}
	x := make([]string, 0)
	return &Instance{avoiding: &x}
}

func (n *Node) StartInstance(challenge Spec) {
	fmt.Println("Starting an instance of", challenge.Name)
	instance, err := n.Client.StartContainer(challenge, n.bindIp)
	if err != nil {
		fmt.Println(err)
		return
	}

	n.mutex.Lock()
	instances, ok := n.challengeInstances[challenge]
	if !ok {
		instances = make([]*Instance, 0)
	}
	n.challengeInstances[challenge] = append(instances, &instance)
	n.mutex.Unlock()
}

func (n *Node) AvoidInstance(user string, instance *Instance) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	avoiding := append(*instance.avoiding, user)
	instance.avoiding = &avoiding
	n.userAvoids[user] = instance
}

func (n *Node) Disconnect(user string) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	instance := n.GetCurrentUserInstance(user)
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

	delete(n.userInstances, user)
}

func (n *Node) StopInstance(instance *Instance) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	_ = n.Client.StopContainer(instance.Container)
	instance.Stopped = true
}
