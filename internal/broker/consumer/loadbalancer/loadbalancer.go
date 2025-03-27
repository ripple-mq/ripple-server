package loadbalancer

import "math/rand"

type ReadReqLoadBalancer struct {
}

func NewReadReqLoadBalancer() *ReadReqLoadBalancer {
	return &ReadReqLoadBalancer{}
}

// TODO: Need to change loadbalancing strategy
func (t *ReadReqLoadBalancer) GetIndex(n int) int {
	return t.randomInt(0, n)
}

func (t *ReadReqLoadBalancer) randomInt(min, max int) int {
	return rand.Intn(max-min+1) + min
}
