package devices

import (
	"context"
	"errors"
	"sync"
)

type Printer interface {
	Name() string
	Print(context.Context, []byte) error
}

type Scanner interface {
	Name() string
	Scan(context.Context) (string, error)
}

type Scale interface {
	Name() string
	WeightGrams(context.Context) (int64, error)
}

type Drawer interface {
	Name() string
	Open(context.Context) error
}

type Display interface {
	Name() string
	Show(context.Context, string) error
}

type Registry struct {
	mu      sync.RWMutex
	printer map[string]Printer
	scanner map[string]Scanner
	scale   map[string]Scale
	drawer  map[string]Drawer
	display map[string]Display
}

func NewRegistry() *Registry {
	return &Registry{
		printer: make(map[string]Printer), scanner: make(map[string]Scanner),
		scale: make(map[string]Scale), drawer: make(map[string]Drawer),
		display: make(map[string]Display),
	}
}

func (r *Registry) RegisterPrinter(device Printer) error {
	return register(r.printer, device.Name(), device, &r.mu)
}

func (r *Registry) RegisterScanner(device Scanner) error {
	return register(r.scanner, device.Name(), device, &r.mu)
}

func (r *Registry) RegisterScale(device Scale) error {
	return register(r.scale, device.Name(), device, &r.mu)
}

func (r *Registry) RegisterDrawer(device Drawer) error {
	return register(r.drawer, device.Name(), device, &r.mu)
}

func (r *Registry) RegisterDisplay(device Display) error {
	return register(r.display, device.Name(), device, &r.mu)
}

func (r *Registry) Printer(name string) (Printer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	device, ok := r.printer[name]
	return device, ok
}

func (r *Registry) Scanner(name string) (Scanner, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	device, ok := r.scanner[name]
	return device, ok
}

func register[T any](devices map[string]T, name string, device T, mu *sync.RWMutex) error {
	if name == "" {
		return errors.New("device and name are required")
	}
	mu.Lock()
	defer mu.Unlock()
	if _, exists := devices[name]; exists {
		return errors.New("device is already registered")
	}
	devices[name] = device
	return nil
}
