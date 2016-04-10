package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/noironetworks/cilium-net/bpf/lxcmap"
	"github.com/noironetworks/cilium-net/cilium-net-daemon/daemon"
	s "github.com/noironetworks/cilium-net/cilium-net-daemon/server"
	common "github.com/noironetworks/cilium-net/common"

	consulAPI "github.com/noironetworks/cilium-net/Godeps/_workspace/src/github.com/hashicorp/consul/api"
	"github.com/noironetworks/cilium-net/Godeps/_workspace/src/github.com/op/go-logging"
)

const(
	RFC3339Mili     = "2006-01-02T15:04:05.999Z07:00"
)

var (
	dockerEndpoint     string
	kubernetesEndpoint string
	socketPath         string
	logLevel           string
	nodeAddrStr        string
	nodeAddr           net.IP
	ipv4Prefix         string
	v4range            string
	ipv4Range          *net.IPNet
	device             string
	libDir             string
	runDir             string
	consulAddr         string
	hostname           string
	disablePolicy      bool
	lxcMap             *lxcmap.LXCMap
	log                = logging.MustGetLogger("cilium-net")
)

func setupLOG(logLevel, hostname string) {

	fileFormat := logging.MustStringFormatter(
		`%{time:` + RFC3339Mili + `} ` + hostname +
		` %{shortfunc} > %{level:.4s} %{id:03x} %{message}`,
	)

	level, err := logging.LogLevel(logLevel)
	if err != nil {
		log.Fatal(err)
	}

	backend := logging.NewLogBackend(os.Stderr, "", 0)
	oBF := logging.NewBackendFormatter(backend, fileFormat)

	backendLeveled := logging.SetBackend(oBF)
	backendLeveled.SetLevel(level, "")
	log.SetBackend(backendLeveled)
}

func initBPF() {
	var args []string

	if err := os.Chdir(runDir); err != nil {
		log.Fatalf("Could not change to runtime directory %s: \"%s\"",
			runDir, err)
		return
	}

	f, err := os.Create("./globals/node_config.h")
	if err != nil {
		log.Warningf("Failed to create node configuration file: %s", err)
		return

	}

	hostIP := make(net.IP, len(nodeAddr))
	copy(hostIP, nodeAddr)
	hostIP[14] = 0xff
	hostIP[15] = 0xff

	fmt.Fprintf(f, ""+
		"/*\n"+
		" * Node-IP: %s\n"+
		" * Host-IP: %s\n"+
		" */\n\n",
		nodeAddr.String(), hostIP.String())

	if logLevel == "debug" {
		f.WriteString("#define DEBUG\n")
	}

	if disablePolicy {
		f.WriteString("#define DISABLE_POCLIY_ENFORCEMENT\n")
	}

	fmt.Fprintf(f, "#define NODE_ID %#x\n", common.NodeAddr2ID(nodeAddr))
	f.WriteString(common.FmtDefineArray("ROUTER_IP", nodeAddr))

	SrcPrefix := net.ParseIP(ipv4Prefix)
	DstPrefix := net.ParseIP(ipv4Prefix)
	f.WriteString(common.FmtDefineAddress("NAT46_SRC_PREFIX", SrcPrefix))
	f.WriteString(common.FmtDefineAddress("NAT46_DST_PREFIX", DstPrefix))

	f.WriteString(common.FmtDefineAddress("HOST_IP", hostIP))

	fmt.Fprintf(f, "#define IPV4_RANGE %#x\n", binary.LittleEndian.Uint32(ipv4Range.IP))
	fmt.Fprintf(f, "#define IPV4_MASK %#x\n", binary.LittleEndian.Uint32(ipv4Range.Mask))

	ipv4Gw := make(net.IP, len(ipv4Range.IP))
	copy(ipv4Gw, ipv4Range.IP)
	ipv4Gw[2] = 0xff
	ipv4Gw[3] = 0xff
	fmt.Fprintf(f, "#define IPV4_GW %#x\n", binary.LittleEndian.Uint32(ipv4Gw))

	f.Close()

	if device != "undefined" {
		args = []string{libDir, nodeAddr.String(), ipv4Range.IP.String(), "direct", device}
	} else {
		args = []string{libDir, nodeAddr.String(), ipv4Range.IP.String(), "vxlan"}
	}

	out, err := exec.Command(libDir+"/init.sh", args...).CombinedOutput()
	if err != nil {
		log.Warningf("Command execution failed: %s", err)
		log.Warningf("Command output:\n%s", out)
		return
	}

	lxcMap, err = lxcmap.OpenMap(common.BPFMap)
	if err != nil {
		log.Warningf("Could not create BPF map '%s': %s", common.BPFMap, err)
		return
	}
}

func init() {
	flag.StringVar(&dockerEndpoint, "e", "unix:///var/run/docker.sock", "Register a listener for docker events on the given endpoint")
	flag.StringVar(&kubernetesEndpoint, "k", "http://127.0.0.1:8080", "Kubernetes endpoint to retrieve metadata information of new started containers")
	flag.StringVar(&logLevel, "l", "info", "Set log level, valid options are (debug|info|warning|error|fatal|panic)")
	flag.StringVar(&socketPath, "s", common.CiliumSock, "Sets the socket path to listen for connections")
	flag.StringVar(&nodeAddrStr, "n", "", "IPv6 address of node, must be in correct format")
	flag.StringVar(&device, "d", "undefined", "Device to snoop on")
	flag.StringVar(&consulAddr, "c", "127.0.0.1:8500", "Consul agent address")
	flag.StringVar(&libDir, "D", "/usr/lib/cilium", "Cilium library directory")
	flag.StringVar(&runDir, "R", "/var/run/cilium", "Runtime data directory")
	flag.StringVar(&ipv4Prefix, "ipv4-mapping", common.DefaultIPv4Prefix, "IPv6 prefix to map IPv4 addresses to")
	flag.StringVar(&v4range, "ipv4-range", "", "IPv6 prefix to map IPv4 addresses to")
	flag.StringVar(&hostname, "hostname", "", "Overwrites hostname's value that will be used for log messages")
	flag.BoolVar(&disablePolicy, "disable-policy", false, "Disable policy enforcement")
	flag.Parse()

	if hostname == "" {
		hostname, _ = os.Hostname()
	}
	setupLOG(logLevel, hostname)

	if nodeAddrStr == "" {
		var err error
		nodeAddrStr, err = common.GenerateV6Prefix()
		if err != nil {
			log.Fatalf("Unable to generate IPv6 prefix: %s\n", err)
			return
		}

		log.Infof("Generated IPv6 prefix: %s\n", nodeAddrStr)
	}

	addr := net.ParseIP(nodeAddrStr)
	if addr == nil {
		log.Fatalf("Invalid node address \"%s\", please specifcy node address using -n", nodeAddrStr)
		return
	}

	if !common.ValidNodeAddress(addr) {
		log.Fatalf("Invalid node address: %s", nodeAddrStr)
	}

	var err error
	nodeAddr, _, err = net.ParseCIDR(addr.String() + "/64")
	if err != nil {
		log.Fatalf("Invalid CIDR %s", addr.String())
		return
	}

	if v4range == "" {
		v4range, err = common.GenerateV4Range()
		if err != nil {
			log.Fatalf("Unable to generate IPv6 prefix: %s\n", err)
			return
		}

		log.Infof("Generated IPv4 range: %s\n", v4range)
	}

	_, ipv4Range, err = net.ParseCIDR(v4range)
	if err != nil {
		log.Fatalf("Invalid IPv4 range %s: %s\n", v4range, err)
		return
	}

	if ones, _ := ipv4Range.Mask.Size(); ones != common.DefaultIPv4Mask {
		log.Fatalf("IPv4 range %s must be of length %d\n", v4range, common.DefaultIPv4Mask)
		return
	}

	if a := net.ParseIP(ipv4Prefix); a == nil || len(a) != net.IPv6len {
		log.Fatalf("Invalid IPv4 prefix %s", ipv4Prefix)
		return
	}

	initBPF()
}

func main() {
	consulDefaultAPI := consulAPI.DefaultConfig()
	consulDefaultAPI.Address = consulAddr
	daemonConf := daemon.Config{
		LibDir:         libDir,
		LXCMap:         lxcMap,
		NodeAddress:    nodeAddr,
		IPv4Range:      ipv4Range,
		ConsulConfig:   consulDefaultAPI,
		DockerEndpoint: dockerEndpoint,
		K8sEndpoint:    kubernetesEndpoint,
		// TODO: Read from a file
		ValidLabelPrefixes: []string{common.GlobalLabelPrefix},
	}

	d, err := daemon.NewDaemon(&daemonConf)
	if err != nil {
		log.Fatalf("Error while creating daemon: %s", err)
	}
	// Register event listener in docker endpoint
	d.ActivateEventListener()
	d.ActivateConsulWatcher(time.Duration(5 * time.Second))

	server, err := s.NewServer(socketPath, d)
	if err != nil {
		log.Fatalf("Error while creating daemon: %s", err)
	}
	defer server.Stop()
	server.Start()
}
