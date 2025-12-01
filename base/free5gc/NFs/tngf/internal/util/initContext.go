package util

import (
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/free5gc/sctp"
	"github.com/free5gc/tngf/internal/logger"
	"github.com/free5gc/tngf/pkg/context"
	"github.com/free5gc/tngf/pkg/factory"
)

var contextLog *logrus.Entry

const RadiusDefaultSecret = "free5GC"

func InitTNGFContext() bool {
	var ok bool
	contextLog = logger.ContextLog

	if factory.TngfConfig.Configuration == nil {
		contextLog.Error("No TNGF configuration found")
		return false
	}

	tngfContext := context.TNGFSelf()

	// TNGF NF information
	tngfContext.NFInfo = factory.TngfConfig.Configuration.TNGFInfo
	if ok = formatSupportedTAList(&tngfContext.NFInfo); !ok {
		return false
	}

	// AMF SCTP addresses
	if len(factory.TngfConfig.Configuration.AMFSCTPAddresses) == 0 {
		contextLog.Error("No AMF specified")
		return false
	} else {
		for _, amfAddress := range factory.TngfConfig.Configuration.AMFSCTPAddresses {
			amfSCTPAddr := new(sctp.SCTPAddr)
			// IP addresses
			for _, ipAddrStr := range amfAddress.IPAddresses {
				if ipAddr, err := net.ResolveIPAddr("ip", ipAddrStr); err != nil {
					contextLog.Errorf("Resolve AMF IP address failed: %+v", err)
					return false
				} else {
					amfSCTPAddr.IPAddrs = append(amfSCTPAddr.IPAddrs, *ipAddr)
				}
			}
			// Port
			if amfAddress.Port == 0 {
				amfSCTPAddr.Port = 38412
			} else {
				amfSCTPAddr.Port = amfAddress.Port
			}
			// Append to context
			tngfContext.AMFSCTPAddresses = append(tngfContext.AMFSCTPAddresses, amfSCTPAddr)
		}
	}

	// IKE bind address
	if factory.TngfConfig.Configuration.IKEBindAddr == "" {
		contextLog.Error("IKE bind address is empty")
		return false
	} else {
		tngfContext.IKEBindAddress = factory.TngfConfig.Configuration.IKEBindAddr
	}

	// Radius bind address
	if factory.TngfConfig.Configuration.RadiusBindAddr == "" {
		contextLog.Error("IKE bind address is empty")
		return false
	} else {
		tngfContext.RadiusBindAddress = factory.TngfConfig.Configuration.RadiusBindAddr
	}

	// IPSec gateway address
	if factory.TngfConfig.Configuration.IPSecGatewayAddr == "" {
		contextLog.Error("IPSec interface address is empty")
		return false
	} else {
		tngfContext.IPSecGatewayAddress = factory.TngfConfig.Configuration.IPSecGatewayAddr
	}

	// GTP bind address
	if factory.TngfConfig.Configuration.GTPBindAddr == "" {
		contextLog.Error("GTP bind address is empty")
		return false
	} else {
		tngfContext.GTPBindAddress = factory.TngfConfig.Configuration.GTPBindAddr
	}

	// TCP port
	if factory.TngfConfig.Configuration.TCPPort == 0 {
		contextLog.Error("TCP port is not defined")
		return false
	} else {
		tngfContext.TCPPort = factory.TngfConfig.Configuration.TCPPort
	}

	// FQDN
	if factory.TngfConfig.Configuration.FQDN == "" {
		contextLog.Error("FQDN is empty")
		return false
	} else {
		tngfContext.FQDN = factory.TngfConfig.Configuration.FQDN
	}

	// Private key
	{
		var keyPath string

		if factory.TngfConfig.Configuration.PrivateKey == "" {
			contextLog.Warn("No private key file path specified, load default key file...")
			keyPath = TngfDefaultKeyPath
		} else {
			keyPath = factory.TngfConfig.Configuration.PrivateKey
		}

		content, err := os.ReadFile(keyPath)
		if err != nil {
			contextLog.Errorf("Cannot read private key data from file: %+v", err)
			return false
		}
		block, _ := pem.Decode(content)
		if block == nil {
			contextLog.Error("Parse pem failed")
			return false
		}
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			contextLog.Warnf("Parse PKCS8 private key failed: %+v", err)
			contextLog.Info("Parse using PKCS1...")

			key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				contextLog.Errorf("Parse PKCS1 pricate key failed: %+v", err)
				return false
			}
		}
		rsaKey, key_ok := key.(*rsa.PrivateKey)
		if !key_ok {
			contextLog.Error("Private key is not an rsa private key")
			return false
		}

		tngfContext.TNGFPrivateKey = rsaKey
	}

	// Certificate authority
	{
		var keyPath string

		if factory.TngfConfig.Configuration.CertificateAuthority == "" {
			contextLog.Warn("No certificate authority file path specified, load default CA certificate...")
			keyPath = TngfDefaultPemPath
		} else {
			keyPath = factory.TngfConfig.Configuration.CertificateAuthority
		}

		// Read .pem
		content, err := os.ReadFile(keyPath)
		if err != nil {
			contextLog.Errorf("Cannot read certificate authority data from file: %+v", err)
			return false
		}
		// Decode pem
		block, _ := pem.Decode(content)
		if block == nil {
			contextLog.Error("Parse pem failed")
			return false
		}
		// Parse DER-encoded x509 certificate
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			contextLog.Errorf("Parse certificate authority failed: %+v", err)
			return false
		}
		// Get sha1 hash of subject public key info
		sha1Hash := sha1.New()
		if _, write_err := sha1Hash.Write(cert.RawSubjectPublicKeyInfo); write_err != nil {
			contextLog.Errorf("Hash function writing failed: %+v", write_err)
			return false
		}

		tngfContext.CertificateAuthority = sha1Hash.Sum(nil)
	}

	// Certificate
	{
		var keyPath string

		if factory.TngfConfig.Configuration.Certificate == "" {
			contextLog.Warn("No certificate file path specified, load default certificate...")
			keyPath = TngfDefaultPemPath
		} else {
			keyPath = factory.TngfConfig.Configuration.Certificate
		}

		// Read .pem
		content, err := os.ReadFile(keyPath)
		if err != nil {
			contextLog.Errorf("Cannot read certificate data from file: %+v", err)
			return false
		}
		// Decode pem
		block, _ := pem.Decode(content)
		if block == nil {
			contextLog.Error("Parse pem failed")
			return false
		}

		tngfContext.TNGFCertificate = block.Bytes
	}

	// Radius Secret
	{
		if factory.TngfConfig.Configuration.RadiusSecret == "" {
			contextLog.Warn("No RADIUS secret specified, load default secret...")
			tngfContext.RadiusSecret = RadiusDefaultSecret
		} else {
			tngfContext.RadiusSecret = factory.TngfConfig.Configuration.RadiusSecret
		}
	}

	// UE IP address range
	if factory.TngfConfig.Configuration.UEIPAddressRange == "" {
		contextLog.Error("UE IP address range is empty")
		return false
	} else {
		_, ueIPRange, err := net.ParseCIDR(factory.TngfConfig.Configuration.UEIPAddressRange)
		if err != nil {
			contextLog.Errorf("Parse CIDR failed: %+v", err)
			return false
		}
		tngfContext.Subnet = ueIPRange
	}

	// XFRM related
	ikeBindIfaceName, err := GetInterfaceName(factory.TngfConfig.Configuration.IKEBindAddr)
	if err != nil {
		contextLog.Error(err)
		return false
	} else {
		tngfContext.XfrmParentIfaceName = ikeBindIfaceName
	}

	if factory.TngfConfig.Configuration.XfrmIfaceName == "" {
		contextLog.Error("XFRM interface Name is empty, set to default \"ipsec\"")
		tngfContext.XfrmIfaceName = "ipsec"
	} else {
		tngfContext.XfrmIfaceName = factory.TngfConfig.Configuration.XfrmIfaceName
	}

	if factory.TngfConfig.Configuration.XfrmIfaceId == 0 {
		contextLog.Warn("XFRM interface id is not defined, set to default value 7")
		tngfContext.XfrmIfaceId = 7
	} else {
		tngfContext.XfrmIfaceId = factory.TngfConfig.Configuration.XfrmIfaceId
	}

	return true
}

func formatSupportedTAList(info *context.TNGFNFInfo) bool {
	for taListIndex := range info.SupportedTAList {
		supportedTAItem := &info.SupportedTAList[taListIndex]

		// Checking TAC
		if supportedTAItem.TAC == "" {
			contextLog.Error("TAC is mandatory.")
			return false
		}
		if len(supportedTAItem.TAC) < 6 {
			contextLog.Trace("Detect configuration TAC length < 6")
			supportedTAItem.TAC = strings.Repeat("0", 6-len(supportedTAItem.TAC)) + supportedTAItem.TAC
			contextLog.Tracef("Changed to %s", supportedTAItem.TAC)
		} else if len(supportedTAItem.TAC) > 6 {
			contextLog.Error("Detect configuration TAC length > 6")
			return false
		}

		// Checking SST and SD
		for plmnListIndex := range supportedTAItem.BroadcastPLMNList {
			broadcastPLMNItem := &supportedTAItem.BroadcastPLMNList[plmnListIndex]

			for sliceListIndex := range broadcastPLMNItem.TAISliceSupportList {
				sliceSupportItem := &broadcastPLMNItem.TAISliceSupportList[sliceListIndex]

				// SST
				if sliceSupportItem.SNSSAI.SST == "" {
					contextLog.Error("SST is mandatory.")
				}
				if len(sliceSupportItem.SNSSAI.SST) < 2 {
					contextLog.Trace("Detect configuration SST length < 2")
					sliceSupportItem.SNSSAI.SST = "0" + sliceSupportItem.SNSSAI.SST
					contextLog.Tracef("Change to %s", sliceSupportItem.SNSSAI.SST)
				} else if len(sliceSupportItem.SNSSAI.SST) > 2 {
					contextLog.Error("Detect configuration SST length > 2")
					return false
				}

				// SD
				if sliceSupportItem.SNSSAI.SD != "" {
					if len(sliceSupportItem.SNSSAI.SD) < 6 {
						contextLog.Trace("Detect configuration SD length < 6")
						sliceSupportItem.SNSSAI.SD = strings.Repeat("0", 6-len(sliceSupportItem.SNSSAI.SD)) + sliceSupportItem.SNSSAI.SD
						contextLog.Tracef("Change to %s", sliceSupportItem.SNSSAI.SD)
					} else if len(sliceSupportItem.SNSSAI.SD) > 6 {
						contextLog.Error("Detect configuration SD length > 6")
						return false
					}
				}
			}
		}
	}

	return true
}

func GetInterfaceName(ipAddress string) (interfaceName string, err error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "nil", err
	}

	res, err := net.ResolveIPAddr("ip4", ipAddress)
	if err != nil {
		return "", fmt.Errorf("error resolving address [%s]: %v", ipAddress, err)
	}

	ipAddress = res.String()

	for _, inter := range interfaces {
		addrs, addr_err := inter.Addrs()
		if addr_err != nil {
			return "nil", addr_err
		}
		for _, addr := range addrs {
			if ipAddress == addr.String()[0:strings.Index(addr.String(), "/")] {
				return inter.Name, nil
			}
		}
	}
	return "", fmt.Errorf("cannot find interface name")
}
