package connector

import (
	"sort"
	"strings"

	"github.com/jennal/goplay/pkg"
)

type HandShakeImpl struct {
	*pkg.HandShakeImpl
	routeMap map[string]bool
}

func NewHandShake() pkg.HandShake {
	return &HandShakeImpl{
		HandShakeImpl: pkg.NewHandShakeImpl(),
		routeMap:      map[string]bool{},
	}
}

func (r *HandShakeImpl) UpdateHandShakeResponse(resp *pkg.HandShakeResponse) {
	r.HandShakeImpl.UpdateHandShakeResponse(resp)
	for route := range r.HandShakeImpl.RpcRoutesMap() {
		if strings.HasPrefix(route, "master.") {
			continue
		}

		if _, ok := r.routeMap[route]; !ok {
			r.routeMap[route] = true
		}
	}

	r.UpdateRoutesMap(r.calcRoutesMap())
}

func (r *HandShakeImpl) calcRoutesMap() map[string]pkg.RouteIndex {
	arr := r.GetKeys()
	m := make(map[string]pkg.RouteIndex)

	for i, str := range arr {
		m[str] = pkg.RouteIndex(i + 1)
	}

	return m
}

func (r *HandShakeImpl) GetKeys() []string {
	arr := []string{}

	for k := range r.routeMap {
		arr = append(arr, k)
	}

	sort.Slice(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})

	return arr
}
