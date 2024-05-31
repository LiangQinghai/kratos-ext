package tarpc

import (
	"context"
	"fmt"
	"github.com/LiangQinghai/kratos-ext/pkg/endpoint"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/lesismal/arpc/extension/micro"
	"strconv"
	"time"
)

type discovery struct {
	w                registry.Watcher
	serviceNamespace string
	serviceManager   micro.ServiceManager
	ctx              context.Context
}

func (d *discovery) watch() {
	for {
		select {
		case <-d.ctx.Done():
			return
		default:

		}
		instances, err := d.w.Next()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Errorf("[resolver] Failed to watch discovery endpoint: %v", err)
			time.Sleep(time.Second)
			continue
		}
		d.update(instances)
	}
}

func (d *discovery) update(serviceInstances []*registry.ServiceInstance) {
	for _, instance := range serviceInstances {
		ept, _ := endpoint.ParseEndpoint(instance.Endpoints, endpoint.Scheme("grpc", false))
		path := fmt.Sprintf("%s/%s/%s", d.serviceNamespace, instance.Name, ept)
		sweight := 10
		if str, ok := instance.Metadata["weight"]; ok {
			if weight, err := strconv.ParseInt(str, 10, 64); err == nil {
				sweight = int(weight)
			}
		}
		d.serviceManager.AddServiceNodes(path, strconv.Itoa(sweight))
	}
}

func (d *discovery) Close() {
	err := d.w.Stop()
	if err != nil {
		log.Errorf("[resolver] failed to watch top: %s", err)
	}
}
