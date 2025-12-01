package handler

import (
	"encoding/binary"
	"errors"

	"github.com/free5gc/aper"
	"github.com/free5gc/nas/nasType"
	"github.com/free5gc/ngap/ngapType"
	"github.com/free5gc/tngf/pkg/radius/message"
)

// 3GPP specified EAP-5G

// Access Network Parameters
type ANParameters struct {
	GUAMI              *ngapType.GUAMI
	SelectedPLMNID     *ngapType.PLMNIdentity
	RequestedNSSAI     *ngapType.AllowedNSSAI
	EstablishmentCause *ngapType.RRCEstablishmentCause
	UEIdentity         *nasType.MobileIdentity5GS
}

func UnmarshalEAP5GData(codedData []byte) (
	eap5GMessageID uint8,
	anParameters *ANParameters,
	nasPDU []byte,
	err error,
) {
	if len(codedData) >= 2 {
		radiusLog.Debug("===== Unmarshal EAP5G Data (Ref: TS24.502 Fig. 9.3.2.2.2-1) =====")

		eap5GMessageID = codedData[0]
		radiusLog.Debugf("Message-Id: %d", eap5GMessageID)
		if eap5GMessageID == message.EAP5GType5GStop {
			return 0, nil, nil, errors.New("eap5GType5GStop")
		}

		codedData = codedData[2:]

		// [TS 24.502 f30] 9.3.2.2.2.3
		// AN-parameter value field in GUAMI, PLMN ID and NSSAI is coded as value part
		// Therefore, IEI of AN-parameter is not needed to be included.
		// anParameter = AN-parameter Type | AN-parameter Length | Value part of IE

		if len(codedData) >= 2 {
			// Length of the AN-Parameter field
			anParameterLength := binary.BigEndian.Uint16(codedData[:2])
			radiusLog.Debugf("AN-parameters length: %d", anParameterLength)

			if anParameterLength != 0 {
				anParameterField := codedData[2:]

				// Bound checking
				if len(anParameterField) < int(anParameterLength) {
					radiusLog.Error("Packet contained error length of value")
					return 0, nil, nil, errors.New("error formatting")
				} else {
					anParameterField = anParameterField[:anParameterLength]
				}

				radiusLog.Debugf("Parsing AN-parameters...: % v", anParameterField)

				anParameters = new(ANParameters)

				// Parse AN-Parameters
				for len(anParameterField) >= 2 {
					parameterType := anParameterField[0]
					// The AN-parameter length field indicates the length of the AN-parameter value field.
					parameterLength := anParameterField[1]

					switch parameterType {
					case message.ANParametersTypeGUAMI:
						// TS 24.502 9.2.1
						radiusLog.Debugf("-> Parameter type: GUAMI")
						if parameterLength != 0 {
							parameterValue := anParameterField[2:]

							if len(parameterValue) < int(parameterLength) {
								return 0, nil, nil, errors.New("error formatting")
							} else {
								parameterValue = parameterValue[:parameterLength]
							}

							if len(parameterValue) != message.ANParametersLenGUAMI {
								return 0, nil, nil, errors.New("unmatched GUAMI length")
							}

							guamiField := make([]byte, 1)
							guamiField = append(guamiField, parameterValue...)
							// Decode GUAMI using aper
							ngapGUAMI := new(ngapType.GUAMI)
							unmarshal_err := aper.UnmarshalWithParams(guamiField, ngapGUAMI, "valueExt")
							if unmarshal_err != nil {
								radiusLog.Errorf("APER unmarshal with parameter failed: %+v", unmarshal_err)
								return 0, nil, nil, errors.New("unmarshal failed when decoding GUAMI")
							}
							anParameters.GUAMI = ngapGUAMI
							radiusLog.Debugf("Unmarshal GUAMI: % x", guamiField)
							radiusLog.Debugf("\tGUAMI: PLMNIdentity[% x], "+
								"AMFRegionID[% x], AMFSetID[% x], AMFPointer[% x]",
								anParameters.GUAMI.PLMNIdentity, anParameters.GUAMI.AMFRegionID,
								anParameters.GUAMI.AMFSetID, anParameters.GUAMI.AMFPointer)
						} else {
							radiusLog.Warn("AN-Parameter GUAMI field empty")
						}
					case message.ANParametersTypeSelectedPLMNID:
						// TS 24.502 9.2.3
						radiusLog.Debugf("-> Parameter type: ANParametersTypeSelectedPLMNID")
						if parameterLength != 0 {
							parameterValue := anParameterField[2:]

							if len(parameterValue) < int(parameterLength) {
								return 0, nil, nil, errors.New("error formatting")
							} else {
								parameterValue = parameterValue[:parameterLength]
							}

							if len(parameterValue) != message.ANParametersLenPLMNID {
								return 0, nil, nil, errors.New("unmatched PLMN ID length")
							}

							plmnField := make([]byte, 1)
							plmnField = append(plmnField, parameterValue...)
							// Decode PLMN using aper
							ngapPLMN := new(ngapType.PLMNIdentity)
							unmarshal_err := aper.UnmarshalWithParams(plmnField, ngapPLMN, "valueExt")
							if unmarshal_err != nil {
								radiusLog.Errorf("APER unmarshal with parameter failed: %v", unmarshal_err)
								return 0, nil, nil, errors.New("unmarshal failed when decoding PLMN")
							}
							anParameters.SelectedPLMNID = ngapPLMN
							radiusLog.Debugf("Unmarshal SelectedPLMNID: % x", plmnField)
							radiusLog.Debugf("\tSelectedPLMNID: % x", anParameters.SelectedPLMNID.Value)
						} else {
							radiusLog.Warn("AN-Parameter PLMN field empty")
						}
					case message.ANParametersTypeRequestedNSSAI:
						// TS 24.501 9.11.3.37
						radiusLog.Debugf("-> Parameter type: ANParametersTypeRequestedNSSAI")
						if parameterLength != 0 {
							parameterValue := anParameterField[2:]

							if len(parameterValue) < int(parameterLength) {
								return 0, nil, nil, errors.New("error formatting")
							} else {
								parameterValue = parameterValue[:parameterLength]
							}

							ngapNSSAI := new(ngapType.AllowedNSSAI)

							// [TS 24501 f30] 9.11.2.8 S-NSSAI
							// s-nssai(LV) consists of
							// len(1 byte) | SST(1) | SD(3,opt) | Mapped HPLMN SST (1,opt) | Mapped HPLMN SD (3,opt)
							// The length of minimum s-nssai comprised of a length and a SST is 2 bytes.

							for len(parameterValue) >= 2 {
								snssaiLength := parameterValue[0]
								snssaiValue := parameterValue[1:]

								if len(snssaiValue) < int(snssaiLength) {
									radiusLog.Error("SNSSAI length error")
									return 0, nil, nil, errors.New("error formatting")
								} else {
									snssaiValue = snssaiValue[:snssaiLength]
								}

								ngapSNSSAIItem := ngapType.AllowedNSSAIItem{}

								if len(snssaiValue) == 1 {
									ngapSNSSAIItem.SNSSAI = ngapType.SNSSAI{
										SST: ngapType.SST{
											Value: []byte{snssaiValue[0]},
										},
									}
								} else if len(snssaiValue) == 4 {
									ngapSNSSAIItem.SNSSAI = ngapType.SNSSAI{
										SST: ngapType.SST{
											Value: []byte{snssaiValue[0]},
										},
										SD: &ngapType.SD{
											Value: []byte{snssaiValue[1], snssaiValue[2], snssaiValue[3]},
										},
									}
								} else {
									radiusLog.Error("Empty SNSSAI value")
									return 0, nil, nil, errors.New("error formatting")
								}

								ngapNSSAI.List = append(ngapNSSAI.List, ngapSNSSAIItem)

								radiusLog.Debugf("Unmarshal SNSSAI: % x", parameterValue[:1+snssaiLength])
								radiusLog.Debugf("\t\t\tSST: % x", ngapSNSSAIItem.SNSSAI.SST.Value)
								sd := ngapSNSSAIItem.SNSSAI.SD
								if sd == nil {
									radiusLog.Debugf("\t\t\tSD: nil")
								} else {
									radiusLog.Debugf("\t\t\tSD: % x", sd.Value)
								}

								// shift parameterValue for parsing next s-nssai
								parameterValue = parameterValue[1+snssaiLength:]
							}
							anParameters.RequestedNSSAI = ngapNSSAI
						} else {
							radiusLog.Warn("AN-Parameter NSSAI is empty")
						}
					case message.ANParametersTypeEstablishmentCause:
						// TS 24.502 9.2.2
						radiusLog.Debugf("-> Parameter type: ANParametersTypeEstablishmentCause")
						if parameterLength != 0 {
							parameterValue := anParameterField[2:]

							if len(parameterValue) < int(parameterLength) {
								return 0, nil, nil, errors.New("error formatting")
							} else {
								parameterValue = parameterValue[:parameterLength]
							}

							if len(parameterValue) != message.ANParametersLenEstCause {
								return 0, nil, nil, errors.New("unmatched Establishment Cause length")
							}

							radiusLog.Debugf("Unmarshal ANParametersTypeEstablishmentCause: % x", parameterValue)

							establishmentCause := parameterValue[0] & 0x0f
							switch establishmentCause {
							case message.EstablishmentCauseEmergency:
								radiusLog.Trace("AN-Parameter establishment cause: Emergency")
							case message.EstablishmentCauseHighPriorityAccess:
								radiusLog.Trace("AN-Parameter establishment cause: High Priority Access")
							case message.EstablishmentCauseMO_Signaling:
								radiusLog.Trace("AN-Parameter establishment cause: MO Signaling")
							case message.EstablishmentCauseMO_Data:
								radiusLog.Trace("AN-Parameter establishment cause: MO Data")
							case message.EstablishmentCauseMPS_PriorityAccess:
								radiusLog.Trace("AN-Parameter establishment cause: MPS Priority Access")
							case message.EstablishmentCauseMCS_PriorityAccess:
								radiusLog.Trace("AN-Parameter establishment cause: MCS Priority Access")
							default:
								radiusLog.Trace("AN-Parameter establishment cause: Unknown. Treat as mo-Data")
								establishmentCause = message.EstablishmentCauseMO_Data
							}

							ngapEstablishmentCause := new(ngapType.RRCEstablishmentCause)
							ngapEstablishmentCause.Value = aper.Enumerated(establishmentCause)

							anParameters.EstablishmentCause = ngapEstablishmentCause
						} else {
							radiusLog.Warn("AN-Parameter establishment cause field empty")
						}
					case message.ANParametersTypeSelectedNID:
						// TS 24.502 9.2.7
						radiusLog.Debugf("-> Parameter type: Selected NID")
						if parameterLength != 0 {
							parameterValue := anParameterField[2:]

							if len(parameterValue) < int(parameterLength) {
								return 0, nil, nil, errors.New("error formatting")
							}
						} else {
							radiusLog.Warn("AN-Parameter selected NID field empty")
						}
					case message.ANParametersTypeUEIdentity:
						// TS 24.501 9.11.3.4
						radiusLog.Debugf("-> Parameter type: UE Identity")
						if parameterLength != 0 {
							parameterValue := anParameterField[2:]

							if len(parameterValue) < int(parameterLength) {
								return 0, nil, nil, errors.New("error formatting")
							} else {
								parameterValue = parameterValue[:parameterLength]
							}

							var iei uint8
							var valLen uint16
							iei = parameterValue[0]
							valLen = binary.BigEndian.Uint16(parameterValue[1:3])

							ueIdentity := nasType.NewMobileIdentity5GS(iei)
							ueIdentity.SetLen(valLen)
							ueIdentity.SetMobileIdentity5GSContents(parameterValue[3:])

							anParameters.UEIdentity = ueIdentity
						} else {
							radiusLog.Warn("AN-Parameter UE Identity field empty")
						}
					default:
						radiusLog.Warn("Unsupported AN-Parameter. Ignore.")
					}

					// shift anParameterField
					anParameterField = anParameterField[2+parameterLength:]
				}
			}

			// shift codedData
			codedData = codedData[2+anParameterLength:]
		} else {
			radiusLog.Error("No AN-Parameter type or length specified")
			return 0, nil, nil, errors.New("error formatting")
		}

		if len(codedData) >= 2 {
			// Length of the NASPDU field
			nasPDULength := binary.BigEndian.Uint16(codedData[:2])
			radiusLog.Debugf("nasPDULength: %d", nasPDULength)

			if nasPDULength != 0 {
				nasPDUField := codedData[2:]

				// Bound checking
				if len(nasPDUField) < int(nasPDULength) {
					return 0, nil, nil, errors.New("error formatting")
				} else {
					nasPDUField = nasPDUField[:nasPDULength]
				}

				radiusLog.Debugf("nasPDUField: % v", nasPDUField)

				nasPDU = append(nasPDU, nasPDUField...)
			} else {
				radiusLog.Error("No NAS PDU included in EAP-5G packet")
				return 0, nil, nil, errors.New("no NAS PDU")
			}
		} else {
			radiusLog.Error("No NASPDU length specified")
			return 0, nil, nil, errors.New("error formatting")
		}

		return eap5GMessageID, anParameters, nasPDU, nil
	} else {
		return 0, nil, nil, errors.New("no data to decode")
	}
}
