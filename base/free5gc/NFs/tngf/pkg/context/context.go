package context

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"math"
	"math/big"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
	gtpv1 "github.com/wmnsk/go-gtp/gtpv1"
	"golang.org/x/net/ipv4"

	"github.com/free5gc/ngap/ngapType"
	"github.com/free5gc/sctp"
	"github.com/free5gc/tngf/internal/logger"
	"github.com/free5gc/util/idgenerator"
)

var contextLog *logrus.Entry

var tngfContext = TNGFContext{}

const RadiusDefaultSecret = "free5GC"

type TNGFContext struct {
	NFInfo           TNGFNFInfo
	AMFSCTPAddresses []*sctp.SCTPAddr

	// ID generator
	RANUENGAPIDGenerator *idgenerator.IDGenerator
	TEIDGenerator        *idgenerator.IDGenerator

	// Pools
	UePool                 sync.Map // map[int64]*TNGFUe, RanUeNgapID as key
	AMFPool                sync.Map // map[string]*TNGFAMF, SCTPAddr as key
	AMFReInitAvailableList sync.Map // map[string]bool, SCTPAddr as key
	IKESA                  sync.Map // map[uint64]*IKESecurityAssociation, SPI as key
	ChildSA                sync.Map // map[uint32]*ChildSecurityAssociation, inboundSPI as key
	GTPConnectionWithUPF   sync.Map // map[string]*gtpv1.UPlaneConn, UPF address as key
	AllocatedUEIPAddress   sync.Map // map[string]*TNGFUe, IPAddr as key
	AllocatedUETEID        sync.Map // map[uint32]*TNGFUe, TEID as key
	RadiusSessionPool      sync.Map // map[string]*RadiusSession, Calling Station ID as key

	// TNGF FQDN
	FQDN string

	// Security data
	CertificateAuthority []byte
	TNGFCertificate      []byte
	TNGFPrivateKey       *rsa.PrivateKey
	RadiusSecret         string

	// UEIPAddressRange
	Subnet *net.IPNet

	// XFRM interface
	XfrmIfaceId         uint32
	XfrmIfaces          sync.Map // map[uint32]*netlink.Link, XfrmIfaceId as key
	XfrmIfaceName       string
	XfrmParentIfaceName string

	// Every UE's first UP IPsec will use default XFRM interface, additoinal UP IPsec will offset its XFRM id
	XfrmIfaceIdOffsetForUP uint32

	// TNGF local address
	IKEBindAddress      string
	RadiusBindAddress   string
	IPSecGatewayAddress string
	GTPBindAddress      string
	TCPPort             uint16

	// TNGF NWt interface IPv4 packet connection
	NWtIPv4PacketConn *ipv4.PacketConn
}

func init() {
	// init log
	contextLog = logger.ContextLog

	// init ID generator
	tngfContext.RANUENGAPIDGenerator = idgenerator.NewGenerator(0, math.MaxInt64)
	tngfContext.TEIDGenerator = idgenerator.NewGenerator(1, math.MaxUint32)
}

// Create new TNGF context
func TNGFSelf() *TNGFContext {
	return &tngfContext
}

func (context *TNGFContext) NewRadiusSession(callingStationID string) *RadiusSession {
	radiusSession := new(RadiusSession)
	radiusSession.CallingStationID = callingStationID
	context.RadiusSessionPool.Store(callingStationID, radiusSession)
	return radiusSession
}

func (context *TNGFContext) DeleteRadiusSession(ranUeNgapId string) {
	context.RadiusSessionPool.Delete(ranUeNgapId)
}

func (context *TNGFContext) RadiusSessionPoolLoad(ranUeNgapId string) (*RadiusSession, bool) {
	ue, ok := context.RadiusSessionPool.Load(ranUeNgapId)
	if ok {
		return ue.(*RadiusSession), ok
	} else {
		return nil, ok
	}
}

func (context *TNGFContext) NewTngfUe() *TNGFUe {
	ranUeNgapId, err := context.RANUENGAPIDGenerator.Allocate()
	if err != nil {
		contextLog.Errorf("New TNGF UE failed: %+v", err)
		return nil
	}
	tngfUe := new(TNGFUe)
	tngfUe.init(ranUeNgapId)
	context.UePool.Store(ranUeNgapId, tngfUe)
	return tngfUe
}

func (context *TNGFContext) DeleteTngfUe(ranUeNgapId int64) {
	context.UePool.Delete(ranUeNgapId)
}

func (context *TNGFContext) UePoolLoad(ranUeNgapId int64) (*TNGFUe, bool) {
	ue, ok := context.UePool.Load(ranUeNgapId)
	if ok {
		return ue.(*TNGFUe), ok
	} else {
		return nil, ok
	}
}

func (context *TNGFContext) NewTngfAmf(sctpAddr string, conn *sctp.SCTPConn) *TNGFAMF {
	amf := new(TNGFAMF)
	amf.init(sctpAddr, conn)
	if item, loaded := context.AMFPool.LoadOrStore(sctpAddr, amf); loaded {
		contextLog.Warn("[Context] NewTngfAmf(): AMF entry already exists.")
		return item.(*TNGFAMF)
	} else {
		return amf
	}
}

func (context *TNGFContext) DeleteTngfAmf(sctpAddr string) {
	context.AMFPool.Delete(sctpAddr)
}

func (context *TNGFContext) AMFPoolLoad(sctpAddr string) (*TNGFAMF, bool) {
	amf, ok := context.AMFPool.Load(sctpAddr)
	if ok {
		return amf.(*TNGFAMF), ok
	} else {
		return nil, ok
	}
}

func (context *TNGFContext) DeleteAMFReInitAvailableFlag(sctpAddr string) {
	context.AMFReInitAvailableList.Delete(sctpAddr)
}

func (context *TNGFContext) AMFReInitAvailableListLoad(sctpAddr string) (bool, bool) {
	flag, ok := context.AMFReInitAvailableList.Load(sctpAddr)
	if ok {
		return flag.(bool), ok
	} else {
		return true, ok
	}
}

func (context *TNGFContext) AMFReInitAvailableListStore(sctpAddr string, flag bool) {
	context.AMFReInitAvailableList.Store(sctpAddr, flag)
}

func (context *TNGFContext) NewIKESecurityAssociation() *IKESecurityAssociation {
	ikeSecurityAssociation := new(IKESecurityAssociation)

	maxSPI := new(big.Int).SetUint64(math.MaxUint64)
	var localSPIuint64 uint64

	for {
		localSPI, err := rand.Int(rand.Reader, maxSPI)
		if err != nil {
			contextLog.Error("[Context] Error occurs when generate new IKE SPI")
			return nil
		}
		localSPIuint64 = localSPI.Uint64()
		if _, duplicate := context.IKESA.LoadOrStore(localSPIuint64, ikeSecurityAssociation); !duplicate {
			break
		}
	}

	ikeSecurityAssociation.LocalSPI = localSPIuint64

	return ikeSecurityAssociation
}

func (context *TNGFContext) DeleteIKESecurityAssociation(spi uint64) {
	context.IKESA.Delete(spi)
}

func (context *TNGFContext) UELoadbyIDi(idi []byte) *TNGFUe {
	var ue *TNGFUe
	context.UePool.Range(func(_, thisUE interface{}) bool {
		strIdi := hex.EncodeToString(idi)
		strSuci := hex.EncodeToString(thisUE.(*TNGFUe).UEIdentity.Buffer)
		contextLog.Debugln("Idi", strIdi)
		contextLog.Debugln("SUCI", strSuci)
		if strIdi == strSuci {
			ue = thisUE.(*TNGFUe)
			return false
		}
		return true
	})
	return ue
}

func (context *TNGFContext) IKESALoad(spi uint64) (*IKESecurityAssociation, bool) {
	securityAssociation, ok := context.IKESA.Load(spi)
	if ok {
		return securityAssociation.(*IKESecurityAssociation), ok
	} else {
		return nil, ok
	}
}

func (context *TNGFContext) DeleteGTPConnection(upfAddr string) {
	context.GTPConnectionWithUPF.Delete(upfAddr)
}

func (context *TNGFContext) GTPConnectionWithUPFLoad(upfAddr string) (*gtpv1.UPlaneConn, bool) {
	conn, ok := context.GTPConnectionWithUPF.Load(upfAddr)
	if ok {
		return conn.(*gtpv1.UPlaneConn), ok
	} else {
		return nil, ok
	}
}

func (context *TNGFContext) GTPConnectionWithUPFStore(upfAddr string, conn *gtpv1.UPlaneConn) {
	context.GTPConnectionWithUPF.Store(upfAddr, conn)
}

func (context *TNGFContext) NewInternalUEIPAddr(ue *TNGFUe) net.IP {
	var ueIPAddr net.IP

	// TODO: Check number of allocated IP to detect running out of IPs
	for {
		ueIPAddr = generateRandomIPinRange(context.Subnet)
		if ueIPAddr != nil {
			if ueIPAddr.String() == context.IPSecGatewayAddress {
				continue
			}
			if _, ok := context.AllocatedUEIPAddress.LoadOrStore(ueIPAddr.String(), ue); !ok {
				break
			}
		}
	}

	return ueIPAddr
}

func (context *TNGFContext) DeleteInternalUEIPAddr(ipAddr string) {
	context.AllocatedUEIPAddress.Delete(ipAddr)
}

func (context *TNGFContext) AllocatedUEIPAddressLoad(ipAddr string) (*TNGFUe, bool) {
	ue, ok := context.AllocatedUEIPAddress.Load(ipAddr)
	if ok {
		return ue.(*TNGFUe), ok
	} else {
		return nil, ok
	}
}

func (context *TNGFContext) NewTEID(ue *TNGFUe) uint32 {
	teid64, err := context.TEIDGenerator.Allocate()
	if err != nil {
		contextLog.Errorf("New TEID failed: %+v", err)
		return 0
	}
	teid32 := uint32(teid64)

	context.AllocatedUETEID.Store(teid32, ue)

	return teid32
}

func (context *TNGFContext) DeleteTEID(teid uint32) {
	context.AllocatedUETEID.Delete(teid)
}

func (context *TNGFContext) AllocatedUETEIDLoad(teid uint32) (*TNGFUe, bool) {
	ue, ok := context.AllocatedUETEID.Load(teid)
	if ok {
		return ue.(*TNGFUe), ok
	} else {
		return nil, ok
	}
}

func (context *TNGFContext) AMFSelection(ueSpecifiedGUAMI *ngapType.GUAMI,
	ueSpecifiedPLMNId *ngapType.PLMNIdentity,
) *TNGFAMF {
	var availableAMF *TNGFAMF
	context.AMFPool.Range(func(key, value interface{}) bool {
		amf := value.(*TNGFAMF)
		if amf.FindAvalibleAMFByCompareGUAMI(ueSpecifiedGUAMI) {
			availableAMF = amf
			return false
		} else {
			// Fail to find through GUAMI served by UE.
			// Try again using SelectedPLMNId
			if amf.FindAvalibleAMFByCompareSelectedPLMNId(ueSpecifiedPLMNId) {
				availableAMF = amf
				return false
			} else {
				return true
			}
		}
	})
	return availableAMF
}

func generateRandomIPinRange(subnet *net.IPNet) net.IP {
	ipAddr := make([]byte, 4)
	randomNumber := make([]byte, 4)

	_, err := rand.Read(randomNumber)
	if err != nil {
		contextLog.Errorf("Generate random number for IP address failed: %+v", err)
		return nil
	}

	// TODO: elimenate network name, gateway, and broadcast
	for i := 0; i < 4; i++ {
		alter := randomNumber[i] & (subnet.Mask[i] ^ 255)
		ipAddr[i] = subnet.IP[i] + alter
	}

	return net.IPv4(ipAddr[0], ipAddr[1], ipAddr[2], ipAddr[3])
}
