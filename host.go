package urknall

import (
	"fmt"
	"github.com/dynport/urknall/fw"
	"strings"
)

// The host type. Use the "NewHost" function to create the basic value.
//
// Please note that you need to set the primary interface (the one the host is accessible on) name, if that is not
// "eth0". That should only be necessary on rare circumstances.
//
//	TODO(gfrey): Add better support for interfaces and IPs.
//	TODO(gfrey): Add handling and support for IPv6 (currently the firewall will block everything).
type Host struct {
	IP        string // Host's IP address used to provision the system.
	User      string // User used to log in.
	Hostname  string // Hostname used on the system.
	Interface string // Primary network interface of the host.

	Paranoid bool // Make the firewall as restrictive as possible.
	WithVPN  bool // Connect host to a VPN. Assumes "tun0" as interface.

	Docker *DockerSettings // Make the host a docker container carrier.

	Rules  []*fw.Rule  // List of rules used for the firewall.
	IPSets []*fw.IPSet // List of ipsets for the firewall.

	packageNames   []string
	userRunlists   []*Runlist
	systemRunlists []*Runlist
}

// If the associated host should run (or build) docker containers this type can be used to configure docker.
type DockerSettings struct {
	Version          string // Docker version to run.
	WithRegistry     bool   // Run an image on this host, that will provide a registry for docker images.
	WithBuildSupport bool   // Configure the associated host so that building images is possible.
	Registry         string // URL of the registry to use.
}

// Get the user used to access the host. If none is given the "root" account is as default.
func (h *Host) user() string {
	if h.User == "" {
		return "root"
	}
	return h.User
}

// Get the host's primary interface. If none is given "eth0" is returned as default.
func (h *Host) publicInterface() string {
	if h.Interface == "" {
		return "eth0"
	}
	return h.Interface
}

func (h *Host) AddCommands(name string, cmds ...interface{}) {
	h.AddPackage(name, NewPackage(cmds...))
}

// Add the given package with the given name to the host.
func (h *Host) AddPackage(name string, pkg Package) {
	if strings.HasPrefix(name, "uk.") {
		panic(fmt.Sprintf(`package name prefix "uk." reserved (in %q)`, name))
	}

	for i := range h.packageNames {
		if h.packageNames[i] == name {
			panic(fmt.Sprintf("package with name %q exists already", name))
		}
	}

	h.packageNames = append(h.packageNames, name)
	h.userRunlists = append(h.userRunlists, &Runlist{name: name, pkg: pkg})
}

// Add the given package with the given name to the host.
func (h *Host) addSystemPackage(name string, pkg Package) (e error) {
	name = "uk." + name
	for i := range h.packageNames {
		if h.packageNames[i] == name {
			return fmt.Errorf("package with name %q exists already", name)
		}
	}

	h.packageNames = append(h.packageNames, name)
	h.systemRunlists = append(h.systemRunlists, &Runlist{name: name, pkg: pkg})
	return nil
}

// Provision the host.
func (h *Host) Provision(opts *ProvisionOptions) (e error) {
	sc := newSSHClient(h, opts)
	return sc.provision()
}

// Provision the given packages into a docker container image tagged with the given tag (the according registry will be
// added automatically). The build will happen on this host, that must be a docker host with build capability.
func (h *Host) CreateDockerImage(baseImage, tag string, pkg Package) (imageId string, e error) {
	if !h.isDockerHost() {
		return "", fmt.Errorf("host %s is not a docker host", h.Hostname)
	}
	dc, e := newDockerClient(h)
	if e != nil {
		return "", e
	}
	return dc.provisionImage(baseImage, tag, pkg)
}

// Get docker version that should be used. Will panic if the host has no docker enabled.
func (h *Host) dockerVersion() string {
	if h.Docker == nil {
		panic("not a docker host")
	}
	if h.Docker.Version == "" {
		return "0.7.0"
	}
	return h.Docker.Version
}

// Predicate to test whether docker must be installed.
func (h *Host) isDockerHost() bool {
	return h.Docker != nil
}

// Predicate to test whether the host should be used to build docker images.
func (h *Host) isDockerBuildHost() bool {
	return h.Docker != nil && h.Docker.WithBuildSupport
}

// Predicate to test whether sudo is required (user for the host is not "root").
func (h *Host) isSudoRequired() bool {
	if h.User != "" && h.User != "root" {
		return true
	}
	return false
}

func (h *Host) runlists() (r []*Runlist) {
	if h.systemRunlists == nil {
		h.buildSystemRunlists()
	}

	r = make([]*Runlist, 0, len(h.systemRunlists)+len(h.userRunlists))
	r = append(r, h.systemRunlists...)
	r = append(r, h.userRunlists...)
	return r
}

func (h *Host) precompileRunlists() (e error) {
	for _, runlist := range h.runlists() {
		if len(runlist.commands) > 0 {
			return fmt.Errorf("pkg %q seems to be packaged already", runlist.name)
		}

		if e = runlist.compile(h); e != nil {
			return e
		}
	}

	return nil
}