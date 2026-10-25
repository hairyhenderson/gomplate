package cidr

import (
	"fmt"
	"math/big"
	"net/netip"

	"go4.org/netipx"
)

// taken from github.com/apparentelymart/go-cidr/ and modified to use the net/netip
// package instead of the stdlib net package - this will hopefully be merged back
// upstream at some point

// SubnetBig takes a parent CIDR range and creates a subnet within it with the
// given number of additional prefix bits and the given network number. It
// differs from Subnet in that it takes a *big.Int for the num, instead of an int.
//
// For example, 10.3.0.0/16, extended by 8 bits, with a network number of 5,
// becomes 10.3.5.0/24 .
func SubnetBig(base netip.Prefix, newBits int, num *big.Int) (netip.Prefix, error) {
	parentLen := base.Bits()
	addrLen := base.Addr().BitLen()

	if newBits < 0 {
		return netip.Prefix{}, fmt.Errorf("can't extend prefix of %d by %d", parentLen, newBits)
	}

	if newBits > addrLen-parentLen {
		return netip.Prefix{}, fmt.Errorf("insufficient address space to extend prefix of %d by %d", parentLen, newBits)
	}

	// the subnet number must fit in newBits bits
	if num.Sign() < 0 || num.BitLen() > newBits {
		return netip.Prefix{}, fmt.Errorf("prefix extension of %d does not accommodate a subnet numbered %d", newBits, num)
	}

	newPrefixLen := parentLen + newBits
	prefix := netip.PrefixFrom(insertNumIntoIP(base.Masked().Addr(), num, newPrefixLen), newPrefixLen)

	return prefix, nil
}

// HostBig takes a parent CIDR range and turns it into a host IP address with
// the given host number. It differs from Host in that it takes a *big.Int for
// the num, instead of an int.
//
// For example, 10.3.0.0/16 with a host number of 2 gives 10.3.0.2.
func HostBig(base netip.Prefix, num *big.Int) (netip.Addr, error) {
	parentLen := base.Bits()
	addrLen := base.Addr().BitLen()

	hostLen := addrLen - parentLen

	// the number of hosts in the network
	size := new(big.Int).Lsh(big.NewInt(1), uint(hostLen))

	// negative numbers count back from the end of the network
	hostNum := new(big.Int).Set(num)
	if hostNum.Sign() < 0 {
		hostNum.Add(hostNum, size)
	}

	if hostNum.Sign() < 0 || hostNum.Cmp(size) >= 0 {
		return netip.Addr{}, fmt.Errorf("prefix of %d does not accommodate a host numbered %d", parentLen, num)
	}

	return insertNumIntoIP(base.Masked().Addr(), hostNum, addrLen), nil
}

func ipToInt(ip netip.Addr) (*big.Int, int) {
	val := &big.Int{}
	val.SetBytes(ip.AsSlice())

	return val, ip.BitLen()
}

func intToIP(ipInt *big.Int, bits int) netip.Addr {
	ipBytes := ipInt.Bytes()
	ret := make([]byte, bits/8)
	// Pack our IP bytes into the end of the return array,
	// since big.Int.Bytes() removes front zero padding.
	for i := 1; i <= len(ipBytes); i++ {
		ret[len(ret)-i] = ipBytes[len(ipBytes)-i]
	}

	addr, ok := netip.AddrFromSlice(ret)
	if !ok {
		panic("invalid IP address")
	}

	return addr
}

func insertNumIntoIP(ip netip.Addr, bigNum *big.Int, prefixLen int) netip.Addr {
	ipInt, totalBits := ipToInt(ip)

	// shift a copy, so the caller's number isn't modified
	shifted := new(big.Int).Lsh(bigNum, uint(totalBits-prefixLen))
	ipInt.Or(ipInt, shifted)
	return intToIP(ipInt, totalBits)
}

// PreviousSubnet returns the subnet of the desired mask in the IP space
// just lower than the start of Prefix provided. If the IP space rolls over
// then the second return value is true
func PreviousSubnet(network netip.Prefix, prefixLen int) (netip.Prefix, bool) {
	previousIP := network.Masked().Addr().Prev()

	previous, err := previousIP.Prefix(prefixLen)
	if err != nil {
		return netip.Prefix{}, false
	}
	if !previous.IsValid() {
		return previous, true
	}

	return previous.Masked(), false
}

// NextSubnet returns the next available subnet of the desired mask size
// starting for the maximum IP of the offset subnet
// If the IP exceeds the maximum IP then the second return value is true
func NextSubnet(network netip.Prefix, prefixLen int) (netip.Prefix, bool) {
	currentLast := netipx.PrefixLastIP(network)

	currentSubnet, err := currentLast.Prefix(prefixLen)
	if err != nil {
		return netip.Prefix{}, false
	}

	last := netipx.PrefixLastIP(currentSubnet).Next()
	next, err := last.Prefix(prefixLen)
	if err != nil {
		return netip.Prefix{}, false
	}
	if !last.IsValid() {
		return next, true
	}
	return next, false
}
