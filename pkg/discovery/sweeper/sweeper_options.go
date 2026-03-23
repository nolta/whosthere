package sweeper

import (
	"errors"
	"net"
	"time"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

type Option func(*Sweeper) error

func WithSweeperInterface(iface *discovery.InterfaceInfo) Option {
	return func(s *Sweeper) error {
		if iface == nil {
			return errors.New("interface cannot be nil")
		}
		s.iface = iface
		return nil
	}
}

// WithSweeperTargetSubnets sets explicit subnets for sweeping.
// When set, the sweeper sweeps these subnets instead of the interface subnet.
func WithSweeperTargetSubnets(subnets []*net.IPNet) Option {
	return func(s *Sweeper) error {
		s.targetSubnets = subnets
		return nil
	}
}

// WithSweeperInterval sets the time between sweep cycles.
// Each sweep triggers ARP for all IPs in the subnet.
// Must be positive.
//
// Default: 5 minutes (discovery.DefaultSweepInterval)
func WithSweeperInterval(interval time.Duration) Option {
	return func(s *Sweeper) error {
		if interval <= 0 {
			return errors.New("sweep interval must be positive")
		}
		s.interval = interval
		return nil
	}
}

// WithSweeperTimeout sets the maximum duration for each sweep cycle.
// If a sweep takes longer, it's canceled and the next one begins.
// Must be positive.
//
// Default: 20 seconds (discovery.DefaultSweepTimeout)
func WithSweeperTimeout(timeout time.Duration) Option {
	return func(s *Sweeper) error {
		if timeout <= 0 {
			return errors.New("sweep timeout must be positive")
		}
		s.timeout = timeout
		return nil
	}
}

// WithSweeperLogger sets a custom logger for the sweeper.
func WithSweeperLogger(logger discovery.Logger) Option {
	return func(s *Sweeper) error {
		if logger == nil {
			return errors.New("logger cannot be nil")
		}
		s.logger = logger
		return nil
	}
}
