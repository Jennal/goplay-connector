package connector

import "github.com/jennal/goplay/pkg"
import "strings"

type HandShakeImpl struct {
	*pkg.HandShakeImpl
	routeMap pkg.RouteMap
}

func NewHandShake() pkg.HandShake {
	return &HandShakeImpl{
		HandShakeImpl: pkg.NewHandShakeImpl(),
		routeMap:      make(pkg.RouteMap),
	}
}

func (r *HandShakeImpl) UpdateHandShakeResponse(resp *pkg.HandShakeResponse) {
	r.HandShakeImpl.UpdateHandShakeResponse(resp)
	for route := range r.HandShakeImpl.RpcRoutesMap() {
		if strings.HasPrefix(route, "master.") {
			continue
		}

		if _, ok := r.routeMap[route]; !ok {
			r.routeMap[route] = pkg.RouteIndex(len(r.routeMap) + 1)
		}
	}

	r.UpdateRoutesMap(r.routeMap)
}

// func (r *HandShakeImpl) ConvertRouteIndexToRpc(idx RouteIndex) RouteIndex {
// 	//do nothing, connector will implement this func
// 	return idx
// }

// func (r *HandShakeImpl) ConvertRouteIndexFromRpc(idx RouteIndex) RouteIndex {
// 	//do nothing, connector will implement this func
// 	return idx
// }
