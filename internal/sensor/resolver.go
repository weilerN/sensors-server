package sensor

import (
	"google.golang.org/grpc/resolver"
)

var (
	constAddr = []string{"localhost:50051", "localhost:50052", "localhost:50053"}
)

type sensorResolverBuilder struct{}

func (*sensorResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &sensorResolver{
		target:     target,
		cc:         cc,
		addrsStore: constAddr, //init with known server addresses
	}
	r.start()
	return r, nil
}
func (*sensorResolverBuilder) Scheme() string { return "" }

type sensorResolver struct {
	target     resolver.Target
	cc         resolver.ClientConn
	addrsStore []string
}

func (r *sensorResolver) start() {
	addrStrs := r.addrsStore //get the const addresses slice
	addrs := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		addrs[i] = resolver.Address{Addr: s}
	}
	err := r.cc.UpdateState(resolver.State{Addresses: addrs})
	if err != nil {
		return
	}
}
func (*sensorResolver) ResolveNow(o resolver.ResolveNowOptions) {}
func (*sensorResolver) Close()                                  {}

func init() {
	resolver.Register(&sensorResolverBuilder{})
}
