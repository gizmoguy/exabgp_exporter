package text

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var testRibDataFile = filepath.Join("testdata", "rib-out.txt")

func testGetTotalLinesInFile(t *testing.T, f string) int {
	file, err := os.Open(f)
	defer func() {
		require.NoError(t, file.Close())
	}()

	require.NoError(t, err)

	s := bufio.NewScanner(file)
	totalLines := 0
	for s.Scan() {
		totalLines++
	}
	return totalLines
}

func TestParseRibTestData(t *testing.T) {
	file, err := os.ReadFile(testRibDataFile)
	require.NoError(t, err)

	totalLines := testGetTotalLinesInFile(t, testRibDataFile)

	ribs, err := RibFromBytes(file)
	require.NoError(t, err)
	require.Equal(t, totalLines, len(ribs))

}

func TestParseRibString(t *testing.T) {
	var testString = `neighbor 127.0.0.1 local-ip 127.0.0.1 local-as 64496 peer-as 64496 router-id 1.1.1.1 family-allowed in-open ipv4 unicast 192.168.88.248/29 next-hop self med 100`
	m, err := RibEntryFromString(testString)
	require.NoError(t, err)
	require.NotNil(t, m)
	require.Equal(t, "127.0.0.1", m.PeerIP)
	require.Equal(t, "64496", m.PeerAS)
	require.Equal(t, "127.0.0.1", m.LocalIP)
	require.Equal(t, "64496", m.LocalAS)
	require.Equal(t, "ipv4", m.AFI)
	require.Equal(t, "unicast", m.SAFI)
	require.Equal(t, "192.168.88.248/29 next-hop self med 100", m.Details)
	require.Equal(t, "ipv4 unicast", m.Family())
}

func TestParseIPv4UnicastFull(t *testing.T) {
	var testString = `neighbor 127.0.0.1 local-ip 127.0.0.1 local-as 64496 peer-as 64496 router-id 1.1.1.1 family-allowed in-open ipv4 unicast 192.168.88.248/29 next-hop self med 100`
	m, err := RibEntryFromString(testString)
	require.NoError(t, err)
	require.NotNil(t, m)
	ipv4, err := m.IPv4Unicast()
	require.NoError(t, err)
	require.NotNil(t, ipv4)
	require.Equal(t, "192.168.88.248/29", ipv4.NLRI)
	require.Equal(t, "self", ipv4.NextHop)
	require.Equal(t, 100, int(ipv4.Attributes.Med))
}

func TestParseIPv4UnicastNoAttributes(t *testing.T) {
	var testString = `neighbor 127.0.0.1 local-ip 127.0.0.1 local-as 64496 peer-as 64496 router-id 1.1.1.1 family-allowed in-open ipv4 unicast 192.168.88.248/29 next-hop self`
	m, err := RibEntryFromString(testString)
	require.NoError(t, err)
	require.NotNil(t, m)
	ipv4, err := m.IPv4Unicast()
	require.NoError(t, err)
	require.NotNil(t, ipv4)
	require.Equal(t, "192.168.88.248/29", ipv4.NLRI)
	require.Equal(t, "self", ipv4.NextHop)
	require.Empty(t, ipv4.Attributes)
}

func TestParseIPv4UnicastFullAttributesNoList(t *testing.T) {
	var testString = `neighbor 127.0.0.1 local-ip 127.0.0.1 local-as 64496 peer-as 64496 router-id 1.1.1.1 family-allowed in-open ipv4 unicast 192.168.88.248/29 next-hop self origin igp as-path [ 30740 ] med 2000 local-preference 100 community 54591:123 originator-id 192.168.22.1 cluster-list [ 3.3.3.3 ] extended-community target:54591:6`
	m, err := RibEntryFromString(testString)
	require.NoError(t, err)
	require.NotNil(t, m)
	ipv4, err := m.IPv4Unicast()
	require.NoError(t, err)
	require.NotNil(t, ipv4)
	require.Equal(t, "192.168.88.248/29", ipv4.NLRI)
	require.Equal(t, "self", ipv4.NextHop)
	require.Equal(t, "igp", ipv4.Attributes.Origin)
	require.Equal(t, "192.168.22.1", ipv4.Attributes.OriginatorID)
	require.Equal(t, 2000, int(ipv4.Attributes.Med))
	require.Equal(t, 100, int(ipv4.Attributes.LocalPreference))
	require.Equal(t, []int{30740}, ipv4.Attributes.ASPath)
	require.Equal(t, []string{"54591:123"}, ipv4.Attributes.Community)
	require.Equal(t, []string{"target:54591:6"}, ipv4.Attributes.ExtendedCommunity)
	require.Equal(t, []string{"3.3.3.3"}, ipv4.Attributes.ClusterList)
}

func TestParseIPv4UnicastFullAttributesList(t *testing.T) {
	var testString = `neighbor 127.0.0.1 local-ip 127.0.0.1 local-as 64496 peer-as 64496 router-id 1.1.1.1 family-allowed in-open ipv4 unicast 192.168.88.248/29 next-hop self origin igp as-path [ 30740 30740 30740 30740 30740 30740 30740 ] med 2000 local-preference 100 community 54591:123 originator-id 192.168.22.1 cluster-list [ 3.3.3.3 192.168.201.1 ] extended-community [ target:54591:6 l2info:19:0:1500:111 ]`
	m, err := RibEntryFromString(testString)
	require.NoError(t, err)
	require.NotNil(t, m)
	ipv4, err := m.IPv4Unicast()
	require.NoError(t, err)
	require.NotNil(t, ipv4)
	require.Equal(t, "192.168.88.248/29", ipv4.NLRI)
	require.Equal(t, "self", ipv4.NextHop)
	require.Equal(t, "igp", ipv4.Attributes.Origin)
	require.Equal(t, "192.168.22.1", ipv4.Attributes.OriginatorID)
	require.Equal(t, 2000, int(ipv4.Attributes.Med))
	require.Equal(t, 100, int(ipv4.Attributes.LocalPreference))
	require.Equal(t, []int{30740, 30740, 30740, 30740, 30740, 30740, 30740}, ipv4.Attributes.ASPath)
	require.Equal(t, []string{"54591:123"}, ipv4.Attributes.Community)
	require.Equal(t, []string{"target:54591:6", "l2info:19:0:1500:111"}, ipv4.Attributes.ExtendedCommunity)
	require.Equal(t, []string{"3.3.3.3", "192.168.201.1"}, ipv4.Attributes.ClusterList)
}

func TestParseIPv6RibString(t *testing.T) {
	var testString = `neighbor 2001::2 local-ip 2001::1 local-as 64496 peer-as 64496 router-id 1.1.1.1 family-allowed in-open ipv6 unicast 2001:db8:1000::/64 next-hop self med 100`
	m, err := RibEntryFromString(testString)
	require.NoError(t, err)
	require.NotNil(t, m)
	require.Equal(t, "2001::2", m.PeerIP)
	require.Equal(t, "64496", m.PeerAS)
	require.Equal(t, "2001::1", m.LocalIP)
	require.Equal(t, "64496", m.LocalAS)
	require.Equal(t, "ipv6", m.AFI)
	require.Equal(t, "unicast", m.SAFI)
	require.Equal(t, "2001:db8:1000::/64 next-hop self med 100", m.Details)
	require.Equal(t, "ipv6 unicast", m.Family())
}

func TestParseIPv6UnicastFull(t *testing.T) {
	var testString = `neighbor 2001::2 local-ip 2001::1 local-as 64496 peer-as 64496 router-id 1.1.1.1 family-allowed in-open ipv6 unicast 2001:db8:1000::/64 next-hop self med 100`
	m, err := RibEntryFromString(testString)
	require.NoError(t, err)
	require.NotNil(t, m)
	ipv6, err := m.IPv6Unicast()
	require.NoError(t, err)
	require.NotNil(t, ipv6)
	require.Equal(t, "2001:db8:1000::/64", ipv6.NLRI)
	require.Equal(t, "self", ipv6.NextHop)
	require.Equal(t, 100, int(ipv6.Attributes.Med))
}

func TestParseIPv6UnicastNoAttributes(t *testing.T) {
	var testString = `neighbor 2001::2 local-ip 2001::1 local-as 64496 peer-as 64496 router-id 1.1.1.1 family-allowed in-open ipv6 unicast 2001:db8:1000::/64 next-hop self`
	m, err := RibEntryFromString(testString)
	require.NoError(t, err)
	require.NotNil(t, m)
	ipv6, err := m.IPv6Unicast()
	require.NoError(t, err)
	require.NotNil(t, ipv6)
	require.Equal(t, "2001:db8:1000::/64", ipv6.NLRI)
	require.Equal(t, "self", ipv6.NextHop)
	require.Empty(t, ipv6.Attributes)
}

func TestRibEntryFromStringReturnsDetailedParseError(t *testing.T) {
	_, err := RibEntryFromString("neighbor 127.0.0.1 broken")
	require.Error(t, err)

	var parseErr *ParseError
	require.True(t, errors.As(err, &parseErr))
	require.Equal(t, "rib parser", parseErr.Parser)
	require.Equal(t, "neighbor 127.0.0.1 broken", parseErr.Input)
	require.Zero(t, parseErr.Line)
	require.Equal(t, "rib parser: unable to parse input: \"neighbor 127.0.0.1 broken\"", err.Error())
}

func TestRibFromBytesReturnsLineNumberInParseError(t *testing.T) {
	data := strings.Join([]string{
		"neighbor 127.0.0.1 local-ip 127.0.0.1 local-as 64496 peer-as 64496 router-id 1.1.1.1 family-allowed in-open ipv4 unicast 192.168.88.248/29 next-hop self",
		"neighbor 127.0.0.1 broken",
	}, "\n")

	_, err := RibFromBytes([]byte(data))
	require.Error(t, err)

	var parseErr *ParseError
	require.True(t, errors.As(err, &parseErr))
	require.Equal(t, 2, parseErr.Line)
	require.Contains(t, err.Error(), "at line 2")
	require.Contains(t, err.Error(), "neighbor 127.0.0.1 broken")
}

func TestIPv4UnicastReturnsDetailedParseError(t *testing.T) {
	m := &RIBMessage{
		AFI:     "ipv4",
		SAFI:    "unicast",
		Details: "192.168.88.248/29 via self",
	}

	_, err := m.IPv4Unicast()
	require.Error(t, err)

	var parseErr *ParseError
	require.True(t, errors.As(err, &parseErr))
	require.Equal(t, "unicast parser", parseErr.Parser)
	require.Equal(t, "192.168.88.248/29 via self", parseErr.Input)
	require.Equal(t, "unicast parser: unable to parse input: \"192.168.88.248/29 via self\"", err.Error())
}
