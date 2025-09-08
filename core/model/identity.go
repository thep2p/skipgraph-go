package model

import "fmt"

// Address contains network address information
type Address struct {
	hostName string
	port     string
}

// NewAddress initializes and returns an instance of Address with the supplied inputs
func NewAddress(hostname string, port string) Address {
	return Address{
		hostName: hostname,
		port:     port,
	}
}

// HostName returns the hostName
func (a Address) HostName() string {
	return a.hostName
}

// Port returns the port
func (a Address) Port() string {
	return a.port
}

// String stringifies an Address
func (a Address) String() string {
	s := fmt.Sprintf("host name: %s, port: %s", a.HostName(), a.Port())
	return s
}

// Identity is a struct that contains the information of a node in the skip graph.
// More specifically, it is the constituent element of the LookupTable.
type Identity struct {
	id        Identifier       // corresponds to numerical id in traditional skip graph.
	memVector MembershipVector // corresponds to name id in traditional skip graph.
	addr      Address          // holds network address like IP.
}

// NewIdentity constructs and returns an Identity.
func NewIdentity(id Identifier, mv MembershipVector, addr Address) Identity {
	i := Identity{}
	i.SetMemVector(mv)
	i.SetAddr(addr)
	i.SetId(id)
	return i
}

// GetIdentifier returns the Identifier field.
func (i Identity) GetIdentifier() Identifier {
	return i.id
}

// GetMembershipVector returns the MembershipVector field.
func (i Identity) GetMembershipVector() MembershipVector {
	return i.memVector
}

// GetAddress returns the Address field.
func (i Identity) GetAddress() Address {
	return i.addr
}

// SetId sets Identifier.
func (i *Identity) SetId(id Identifier) {
	// TODO validation of the id may be needed.
	i.id = id
}

// SetMemVector sets membershipVector.
func (i *Identity) SetMemVector(mv MembershipVector) {
	// TODO validation of the membershipVector mv may be needed.
	i.memVector = mv
}

// SetAddr sets address.
func (i *Identity) SetAddr(addr Address) {
	// TODO validation of the addr may be needed.
	i.addr = addr
}
