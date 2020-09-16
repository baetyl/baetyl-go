package native

import (
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/utils"
)

type PortAllocator struct {
	base   int
	size   int
	offset int
}

func NewPortAllocator(start, end int) (*PortAllocator, error) {
	if start < 1024 || end > 65535 || start >= end {
		return nil, errors.Errorf("port range (%d) - (%d) is not valid", start, end)
	}
	return &PortAllocator{
		base: start,
		size: end - start + 1,
	}, nil
}

func (p *PortAllocator) Allocate() (int, error) {
	var times int
	for {
		if times == p.size {
			return 0, errors.Errorf("no available ports in range %d-%d", p.base, p.base+p.size-1)
		}
		port := p.base + p.offset
		p.offset++
		p.offset = p.offset % p.size
		if utils.CheckPortAvailable("127.0.0.1", port) {
			return port, nil
		}
		times++
	}
}
