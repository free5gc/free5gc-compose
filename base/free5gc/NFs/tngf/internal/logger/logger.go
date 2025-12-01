package logger

import (
	"github.com/sirupsen/logrus"

	logger_util "github.com/free5gc/util/logger"
)

var (
	Log        *logrus.Logger
	NfLog      *logrus.Entry
	MainLog    *logrus.Entry
	InitLog    *logrus.Entry
	CfgLog     *logrus.Entry
	ContextLog *logrus.Entry
	NgapLog    *logrus.Entry
	IKELog     *logrus.Entry
	RadiusLog  *logrus.Entry
	GTPLog     *logrus.Entry
	NWtCPLog   *logrus.Entry
	NWtUPLog   *logrus.Entry
	RelayLog   *logrus.Entry
	UtilLog    *logrus.Entry
)

func init() {
	fieldsOrder := []string{
		logger_util.FieldNF,
		logger_util.FieldCategory,
	}
	Log = logger_util.New(fieldsOrder)
	NfLog = Log.WithField(logger_util.FieldNF, "TNGF")
	MainLog = NfLog.WithField(logger_util.FieldCategory, "Main")
	InitLog = NfLog.WithField(logger_util.FieldCategory, "Init")
	CfgLog = NfLog.WithField(logger_util.FieldCategory, "CFG")
	ContextLog = NfLog.WithField(logger_util.FieldCategory, "Context")
	NgapLog = NfLog.WithField(logger_util.FieldCategory, "NGAP")
	IKELog = NfLog.WithField(logger_util.FieldCategory, "IKE")
	RadiusLog = NfLog.WithField(logger_util.FieldCategory, "Radius")
	GTPLog = NfLog.WithField(logger_util.FieldCategory, "GTP")
	NWtCPLog = NfLog.WithField(logger_util.FieldCategory, "NWtCP")
	NWtUPLog = NfLog.WithField(logger_util.FieldCategory, "NWtUP")
	RelayLog = NfLog.WithField(logger_util.FieldCategory, "Relay")
	UtilLog = NfLog.WithField(logger_util.FieldCategory, "Util")
}
