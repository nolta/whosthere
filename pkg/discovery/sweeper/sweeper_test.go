package sweeper

import (
	"net"
	"testing"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/stretchr/testify/require"
)

func TestIncrementIP_Carry(t *testing.T) {
	ip := net.IPv4(192, 168, 0, 255).To4()
	next := incrementIP(ip)
	require.Equal(t, "192.168.1.0", next.String())
}

func TestSweeper_GenerateSubnetIPs_SkipsLocalAndIncludesNetworkAndBroadcast(t *testing.T) {
	local := net.IPv4(192, 168, 1, 1).To4()
	_, subnet, err := net.ParseCIDR("192.168.1.1/30")
	require.NoError(t, err)

	s := &Sweeper{logger: &discovery.NoOpLogger{}}
	ips := s.generateSubnetIPs(subnet, local)

	got := make([]string, 0, len(ips))
	for _, ip := range ips {
		got = append(got, ip.String())
	}
	require.Equal(t, []string{"192.168.1.0", "192.168.1.2", "192.168.1.3"}, got)
}

func TestSweeper_GenerateSubnetIPs_IPv6SubnetReturnsEmpty(t *testing.T) {
	_, subnet, err := net.ParseCIDR("2001:db8::/64")
	require.NoError(t, err)

	s := &Sweeper{logger: &discovery.NoOpLogger{}}
	ips := s.generateSubnetIPs(subnet, net.ParseIP("2001:db8::1"))
	require.Empty(t, ips)
}

func TestSweeper_GenerateSubnetIPs_LimitsLargeSubnetTo16(t *testing.T) {
	_, subnet, err := net.ParseCIDR("10.0.0.0/8")
	require.NoError(t, err)

	s := &Sweeper{logger: &discovery.NoOpLogger{}}
	ips := s.generateSubnetIPs(subnet, net.IPv4(10, 0, 0, 1).To4())

	require.Len(t, ips, 65535)
	require.Equal(t, "10.0.0.0", ips[0].String())
	require.Equal(t, "10.0.255.255", ips[len(ips)-1].String())
}

func TestSweeper_SweepAll_UsesTargetSubnetsWhenConfigured(t *testing.T) {
	_, subnet1, _ := net.ParseCIDR("10.0.0.0/30")
	_, subnet2, _ := net.ParseCIDR("10.0.1.0/30")

	localIP := net.IPv4(192, 168, 0, 1).To4()

	s := &Sweeper{
		logger:        &discovery.NoOpLogger{},
		targetSubnets: []*net.IPNet{subnet1, subnet2},
		iface: &discovery.InterfaceInfo{
			IPv4Addr: &localIP,
			IPv4Net:  &net.IPNet{IP: net.IPv4(192, 168, 0, 0), Mask: net.CIDRMask(24, 32)},
		},
	}

	s.sweepAll(t.Context(), localIP)
}

func TestSweeper_SweepAll_FallsBackToInterfaceSubnet(t *testing.T) {
	localIP := net.IPv4(192, 168, 0, 1).To4()
	_, ifaceNet, _ := net.ParseCIDR("192.168.0.0/30")

	s := &Sweeper{
		logger: &discovery.NoOpLogger{},
		iface: &discovery.InterfaceInfo{
			IPv4Addr: &localIP,
			IPv4Net:  ifaceNet,
		},
	}

	s.sweepAll(t.Context(), localIP)
}

func TestSweeper_WithTargetSubnets_Option(t *testing.T) {
	localIP := net.IPv4(192, 168, 0, 1).To4()
	_, ifaceNet, _ := net.ParseCIDR("192.168.0.0/24")
	_, target, _ := net.ParseCIDR("10.0.0.0/24")

	sw, err := New(
		WithSweeperInterface(&discovery.InterfaceInfo{
			Interface: &net.Interface{Name: "test0"},
			IPv4Addr:  &localIP,
			IPv4Net:   ifaceNet,
		}),
		WithSweeperTargetSubnets([]*net.IPNet{target}),
	)
	require.NoError(t, err)
	require.Len(t, sw.targetSubnets, 1)
	require.Equal(t, "10.0.0.0/24", sw.targetSubnets[0].String())
}
