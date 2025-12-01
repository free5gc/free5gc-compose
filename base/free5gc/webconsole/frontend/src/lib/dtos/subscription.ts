import { z } from "zod";
import {
  AuthenticationSubscription,
  ChargingData,
  DnnConfiguration,
  FlowRules,
  Nssai,
  QosFlows,
  SessionManagementSubscriptionData,
  SubscribedUeAmbr,
  Subscription,
  UpSecurity,
} from "../../api";
import { DEFAULT_5QI } from "../const";

export const subscriberAuthDTOSchema = z.object({
  authenticationManagementField: z.string().regex(/^[A-Fa-f0-9]{4}$/), // 16 bit hex string
  authenticationMethod: z.enum(["5G_AKA", "EAP_AKA_PRIME"]),
  sequenceNumber: z.string().regex(/^[A-Fa-f0-9]{12}$/), // 48 bit hex string
  permanentKey: z.string(),
  operatorCodeType: z.enum(["OP", "OPc"]),
  operatorCode: z.string(),
})

interface SubscriberAuthDTO {
  authenticationManagementField: string;
  authenticationMethod: string;
  sequenceNumber: string;
  permanentKey: string;
  operatorCodeType: "OP" | "OPc";
  operatorCode: string;
}

export const ambrDTOSchema = z.object({
  uplink: z.string(),
  downlink: z.string(),
})

interface AmbrDTO {
  uplink: string;
  downlink: string;
}

export const chargingDataDTOSchema = z.object({
  chargingMethod: z.enum(["Online", "Offline"]),
  quota: z.string(),
  unitCost: z.string(),
})

interface ChargingDataDTO {
  chargingMethod: "Online" | "Offline";
  quota: string;
  unitCost: string;
}

export const flowRulesDTOSchema = z.object({
  filter: z.string(),
  precedence: z.number(),
  "5qi": z.number(),
  gbrUL: z.string(),
  gbrDL: z.string(),
  mbrUL: z.string(),
  mbrDL: z.string(),
  chargingData: chargingDataDTOSchema,
})

interface FlowRulesDTO {
  filter: string;
  precedence: number;
  "5qi": number;
  gbrUL: string;
  gbrDL: string;
  mbrUL: string;
  mbrDL: string;
  chargingData: ChargingDataDTO;
}

export const upSecurityDTOSchema = z.object({
  upIntegr: z.string(),
  upConfid: z.string(),
})

interface UpSecurityDTO {
  upIntegr: string;
  upConfid: string;
}

export const DnnConfigurationDTOSchema = z.object({
  default5qi: z.number(),
  sessionAmbr: ambrDTOSchema,
  enableStaticIpv4Address: z.boolean(),
  staticIpv4Address: z.string().optional(),
  flowRules: z.array(flowRulesDTOSchema),
  upSecurity: upSecurityDTOSchema.optional(), 
})

interface DnnConfigurationDTO {
  default5qi: number;
  sessionAmbr: AmbrDTO;
  enableStaticIpv4Address: boolean;
  staticIpv4Address?: string;
  flowRules: FlowRulesDTO[];
  upSecurity?: UpSecurityDTO;
}


export const SnssaiConfigurationDTOSchema = z.object({
  sst: z.number(),
  sd: z.string().optional(),
  isDefault: z.boolean(),
  chargingData: chargingDataDTOSchema,
  dnnConfigurations: z.record(DnnConfigurationDTOSchema),
});

interface SnssaiConfigurationDTO {
  sst: number;
  sd: string;
  isDefault: boolean;
  chargingData: ChargingDataDTO;
  dnnConfigurations: { [key: string]: DnnConfigurationDTO };
}


export const subscriptionDTOSchema = z.object({
  userNumber: z.number().positive(),
  ueId: z.string().length(20).startsWith("imsi-"),
  plmnID: z.string().min(5).max(6),
  gpsi: z.string().optional(),
  auth: subscriberAuthDTOSchema,
  subscribedUeAmbr: ambrDTOSchema,
  SnssaiConfigurations: z.array(SnssaiConfigurationDTOSchema),
})

interface SubscriptionDTO {
  userNumber: number;
  ueId: string;
  plmnID: string;
  gpsi?: string;
  auth: SubscriberAuthDTO;
  subscribedUeAmbr: AmbrDTO;
  SnssaiConfigurations: SnssaiConfigurationDTO[];
}

interface FlowsDTO {
  flowRules: FlowRules[];
  qosFlows: QosFlows[];
  chargingDatas: ChargingData[];
}

interface FlowsMapper {
  map(subscription: SubscriptionDTO): FlowsDTO;
}

class FlowsMapperImpl implements FlowsMapper {
  refNumber: number = 1;
  flowRules: FlowRules[] = [];
  qosFlows: QosFlows[] = [];
  chargingDatas: ChargingData[] = [];

  private buildDnns(subscription: SubscriptionDTO): {
    snssai: string;
    dnn: string;
    sliceCharingData: ChargingDataDTO;
    flowRules: FlowRulesDTO[];
  }[] {
    return subscription.SnssaiConfigurations.reduce(
      (acc, s) => {
        const snssai = s.sst.toString().padStart(2, "0") + s.sd;
        const dnns = Object.entries(s.dnnConfigurations).map(([dnn, dnnConfig]) => ({
          snssai: snssai,
          dnn: dnn,
          sliceCharingData: s.chargingData,
          flowRules: dnnConfig.flowRules,
        }));
        dnns.forEach((dnn) => acc.push(dnn));
        return acc;
      },
      [] as {
        snssai: string;
        dnn: string;
        sliceCharingData: ChargingDataDTO;
        flowRules: FlowRulesDTO[];
      }[],
    );
  }

  map(subscription: SubscriptionDTO): FlowsDTO {
    const dnns = this.buildDnns(subscription);

    dnns.forEach((dnn) => {
      const snssai = dnn.snssai;

      this.chargingDatas.push({
        ...dnn.sliceCharingData,
        snssai: snssai,
        dnn: "",
        filter: "",
      });

      dnn.flowRules.forEach((flow) => {
        const qosRef = this.refNumber++;

        this.flowRules.push({
          filter: flow.filter,
          precedence: flow.precedence,
          snssai: snssai,
          dnn: dnn.dnn,
          qosRef,
        });

        this.qosFlows.push({
          snssai: snssai,
          dnn: dnn.dnn,
          qosRef,
          "5qi": flow["5qi"],
          mbrUL: flow.mbrUL,
          mbrDL: flow.mbrDL,
          gbrUL: flow.gbrUL,
          gbrDL: flow.gbrDL,
        });

        this.chargingDatas.push({
          ...flow.chargingData,
          snssai: snssai,
          dnn: dnn.dnn,
          filter: flow.filter,
          qosRef,
        });
      });
    });

    return {
      flowRules: this.flowRules,
      qosFlows: this.qosFlows,
      chargingDatas: this.chargingDatas,
    };
  }
}

interface SubscriptionMapper {
  mapFromDto(subscription: SubscriptionDTO): Subscription;
  mapFromSubscription(subscription: Subscription): SubscriptionDTO;
}

class SubscriptionMapperImpl implements SubscriptionMapper {
  constructor(private readonly flowsBuilder: FlowsMapper) {}

  mapFromDto(subscription: SubscriptionDTO): Subscription {
    const flows = this.flowsBuilder.map(subscription);

    return {
      userNumber: subscription.userNumber,
      ueId: subscription.ueId,
      plmnID: subscription.plmnID,
      AuthenticationSubscription: this.buildSubscriberAuth(subscription.auth),
      AccessAndMobilitySubscriptionData: {
        gpsis: [`msisdn-${subscription.gpsi ?? ""}`],
        subscribedUeAmbr: this.buildSubscriberAmbr(subscription.subscribedUeAmbr),
        nssai: {
          defaultSingleNssais: subscription.SnssaiConfigurations.filter((s) => s.isDefault).map(
            (s) => this.buildNssai(s),
          ),
          singleNssais: subscription.SnssaiConfigurations.filter((s) => !s.isDefault).map((s) =>
            this.buildNssai(s),
          ),
        },
      },

      SessionManagementSubscriptionData: subscription.SnssaiConfigurations.map((s) =>
        this.buildSessionManagementSubscriptionData(s),
      ),

      SmfSelectionSubscriptionData: {
        subscribedSnssaiInfos: Object.fromEntries(
          subscription.SnssaiConfigurations.map((s) => [
            s.sst.toString().padStart(2, "0") + s.sd,
            {
              dnnInfos: Object.keys(s.dnnConfigurations).map((dnn) => ({
                dnn: dnn,
              })),
            },
          ]),
        ),
      },

      AmPolicyData: {
        subscCats: ["free5gc"],
      },

      SmPolicyData: {
        smPolicySnssaiData: Object.fromEntries(
          subscription.SnssaiConfigurations.map((s) => [
            s.sst.toString().padStart(2, "0") + s.sd,
            {
              snssai: this.buildNssai(s),
              smPolicyDnnData: Object.fromEntries(
                Object.keys(s.dnnConfigurations).map((dnn) => [dnn, { dnn: dnn }]),
              ),
            },
          ]),
        ),
      },

      FlowRules: flows.flowRules,
      QosFlows: flows.qosFlows,
      ChargingDatas: flows.chargingDatas,
    };
  }

  mapFromSubscription(subscription: Subscription): SubscriptionDTO {
    return {
      userNumber: 1,
      ueId: subscription.ueId,
      plmnID: subscription.plmnID,
      gpsi: subscription.AccessAndMobilitySubscriptionData.gpsis?.[0].slice(7) ?? "",
      auth: {
        authenticationManagementField:
          subscription.AuthenticationSubscription.authenticationManagementField ?? "",
        authenticationMethod: subscription.AuthenticationSubscription.authenticationMethod,
        sequenceNumber: subscription.AuthenticationSubscription.sequenceNumber,
        permanentKey: subscription.AuthenticationSubscription.permanentKey.permanentKeyValue,
        operatorCodeType: subscription.AuthenticationSubscription.milenage?.op?.opValue
          ? "OP"
          : "OPc",
        operatorCode: subscription.AuthenticationSubscription.milenage?.op?.opValue
          ? subscription.AuthenticationSubscription.milenage.op.opValue
          : subscription.AuthenticationSubscription.opc?.opcValue ?? "",
      },
      subscribedUeAmbr: {
        uplink: subscription.AccessAndMobilitySubscriptionData.subscribedUeAmbr?.uplink ?? "",
        downlink: subscription.AccessAndMobilitySubscriptionData.subscribedUeAmbr?.downlink ?? "",
      },
      SnssaiConfigurations: subscription.SessionManagementSubscriptionData.map((s) => ({
        sst: s.singleNssai.sst,
        sd: s.singleNssai.sd ?? "",
        isDefault: this.snssaiIsDefault(s.singleNssai, subscription),
        chargingData: this.findSliceChargingData(s.singleNssai, subscription),
        dnnConfigurations: Object.fromEntries(
          Object.entries(s.dnnConfigurations ?? {}).map(([key, value]) => [
            key,
            {
              default5qi: value["5gQosProfile"]?.["5qi"] ?? DEFAULT_5QI,
              sessionAmbr: {
                uplink: value.sessionAmbr?.uplink ?? "",
                downlink: value.sessionAmbr?.downlink ?? "",
              },
              enableStaticIpv4Address: value.staticIpAddress?.length !== 0,
              staticIpv4Address: value.staticIpAddress?.[0]?.ipv4Addr ?? "",
              flowRules: this.parseDnnFlowRules(s.singleNssai, key, subscription),
              upSecurity: value.upSecurity,
            } satisfies DnnConfigurationDTO,
          ]),
        ),
      })),
    };
  }

  private snssaiIsDefault(nssai: Nssai, subscription: Subscription): boolean {
    return (
      subscription.AccessAndMobilitySubscriptionData.nssai?.defaultSingleNssais?.some(
        (n) => n.sst === nssai.sst && n.sd === nssai.sd,
      ) ?? false
    );
  }

  private findSliceChargingData(nssai: Nssai, subscription: Subscription): ChargingDataDTO {
    const charingData = subscription.ChargingDatas.find((c) => {
      if (c.dnn !== "" || c.filter !== "") {
        return false;
      }

      return c.snssai === nssai.sst.toString().padStart(2, "0") + nssai.sd;
    });

    return {
      chargingMethod: charingData?.chargingMethod === "Online" ? "Online" : "Offline",
      quota: charingData?.quota ?? "",
      unitCost: charingData?.unitCost ?? "",
    };
  }

  private parseDnnFlowRules(
    snssai: Nssai,
    dnn: string,
    subscription: Subscription,
  ): FlowRulesDTO[] {
    const qosFlows = subscription.QosFlows.filter(
      (f) => f.dnn === dnn && f.snssai === snssai.sst.toString().padStart(2, "0") + snssai.sd,
    );

    return qosFlows.map((f) => {
      const flowRule = subscription.FlowRules.find((r) => r.qosRef === f.qosRef);
      const chargingData = subscription.ChargingDatas.find((c) => c.qosRef === f.qosRef);

      return {
        filter: flowRule?.filter ?? "",
        precedence: flowRule?.precedence ?? 0,
        "5qi": f["5qi"] ?? DEFAULT_5QI,
        gbrUL: f.gbrUL ?? "",
        gbrDL: f.gbrDL ?? "",
        mbrUL: f.mbrUL ?? "",
        mbrDL: f.mbrDL ?? "",
        chargingData: {
          chargingMethod: chargingData?.chargingMethod === "Online" ? "Online" : "Offline",
          quota: chargingData?.quota ?? "",
          unitCost: chargingData?.unitCost ?? "",
        },
      }
    });
  }

  private buildSubscriberAuth(data: SubscriberAuthDTO): AuthenticationSubscription {
    return {
      authenticationMethod: data.authenticationMethod,
      permanentKey: {
        permanentKeyValue: data.permanentKey,
        encryptionKey: 0,
        encryptionAlgorithm: 0,
      },
      sequenceNumber: data.sequenceNumber,
      authenticationManagementField: data.authenticationManagementField,
      milenage:
        data.operatorCodeType === "OP"
          ? {
              op: {
                opValue: data.operatorCode,
                encryptionKey: 0,
                encryptionAlgorithm: 0,
              },
            }
          : { op: { opValue: "", encryptionKey: 0, encryptionAlgorithm: 0 } },
      opc:
        data.operatorCodeType === "OPc"
          ? { opcValue: data.operatorCode, encryptionKey: 0, encryptionAlgorithm: 0 }
          : { opcValue: "", encryptionKey: 0, encryptionAlgorithm: 0 },
    };
  }

  buildSubscriberAmbr(data: AmbrDTO): SubscribedUeAmbr {
    return {
      uplink: data.uplink,
      downlink: data.downlink,
    };
  }

  buildNssai(data: SnssaiConfigurationDTO): Nssai {
    return {
      sst: data.sst,
      sd: data.sd,
    };
  }

  buildSessionManagementSubscriptionData(
    data: SnssaiConfigurationDTO,
  ): SessionManagementSubscriptionData {
    return {
      singleNssai: {
        sst: data.sst,
        sd: data.sd,
      },
      dnnConfigurations: Object.fromEntries(
        Object.entries(data.dnnConfigurations).map(([key, value]) => [
          key,
          this.buildDnnConfiguration(value),
        ]),
      ),
    };
  }

  buildDnnConfiguration(data: DnnConfigurationDTO): DnnConfiguration {
    return {
      pduSessionTypes: {
        defaultSessionType: "IPV4",
        allowedSessionTypes: ["IPV4"],
      },
      sscModes: {
        defaultSscMode: "SSC_MODE_1",
        allowedSscModes: ["SSC_MODE_2", "SSC_MODE_3"],
      },
      "5gQosProfile": {
        "5qi": data.default5qi,
        arp: {
          priorityLevel: 8,
          preemptCap: "",
          preemptVuln: "",
        },
        priorityLevel: 8,
      },
      sessionAmbr: this.buildSubscriberAmbr(data.sessionAmbr),
      staticIpAddress: data.enableStaticIpv4Address ? [{ ipv4Addr: data.staticIpv4Address }] : [],
      upSecurity: this.buildUpSecurity(data.upSecurity),
    };
  }

  buildUpSecurity(data: UpSecurityDTO | undefined): UpSecurity | undefined {
    if (!data) {
      return undefined;
    }

    return {
      upIntegr: data.upIntegr,
      upConfid: data.upConfid,
    };
  }
}

export const defaultSubscriptionDTO = (): SubscriptionDTO => ({
  userNumber: 1,
  ueId: "imsi-208930000000001",
  plmnID: "20893",
  gpsi: "",
  auth: {
    authenticationManagementField: "8000",
    authenticationMethod: "5G_AKA",
    sequenceNumber: "000000000023",
    permanentKey: "8baf473f2f8fd09487cccbd7097c6862",
    operatorCodeType: "OPc",
    operatorCode: "8e27b6af0e692e750f32667a3b14605d",
  },
  subscribedUeAmbr: {
    uplink: "1 Gbps",
    downlink: "2 Gbps",
  },
  SnssaiConfigurations: [
    {
      sst: 1,
      sd: "010203",
      isDefault: true,
      chargingData: {
        chargingMethod: "Offline",
        quota: "100000",
        unitCost: "1",
      },
      dnnConfigurations: {
        internet: {
          enableStaticIpv4Address: false,
          default5qi: 9,
          sessionAmbr: {
            uplink: "1000 Mbps",
            downlink: "1000 Mbps",
          },
          staticIpv4Address: "",
          flowRules: [
            {
              filter: "1.1.1.1/32",
              precedence: 128,
              "5qi": 8,
              gbrUL: "108 Mbps",
              gbrDL: "108 Mbps",
              mbrUL: "208 Mbps",
              mbrDL: "208 Mbps",
              chargingData: {
                chargingMethod: "Offline",
                quota: "100000",
                unitCost: "1",
              },
            },
          ],
        },
      },
    },
    {
      sst: 1,
      sd: "112233",
      isDefault: false,
      chargingData: {
        chargingMethod: "Online",
        quota: "100000",
        unitCost: "1",
      },
      dnnConfigurations: {
        internet: {
          enableStaticIpv4Address: false,
          default5qi: 8,
          sessionAmbr: {
            uplink: "1000 Mbps",
            downlink: "1000 Mbps",
          },
          staticIpv4Address: "",
          flowRules: [
            {
              filter: "1.1.1.1/32",
              precedence: 127,
              "5qi": 7,
              gbrUL: "207 Mbps",
              gbrDL: "207 Mbps",
              mbrUL: "407 Mbps",
              mbrDL: "407 Mbps",
              chargingData: {
                chargingMethod: "Online",
                quota: "5000",
                unitCost: "1",
              },
            },
          ],
        },
      },
    },
  ],
});

export const defaultSnssaiConfiguration = (): SnssaiConfigurationDTO => ({
  sst: 1,
  sd: "",
  isDefault: false,
  chargingData: {
    chargingMethod: "Offline",
    quota: "100000",
    unitCost: "1",
  },
  dnnConfigurations: { internet: defaultDnnConfig() },
});

export const defaultDnnConfig = (): DnnConfigurationDTO => ({
  default5qi: DEFAULT_5QI,
  sessionAmbr: {
    uplink: "1000 Mbps",
    downlink: "1000 Mbps",
  },
  enableStaticIpv4Address: false,
  staticIpv4Address: "",
  flowRules: [defaultFlowRule()],
  upSecurity: undefined,
});

export const defaultFlowRule = (): FlowRulesDTO => ({
  filter: "1.1.1.1/32",
  precedence: 128,
  "5qi": 9,
  gbrUL: "208 Mbps",
  gbrDL: "208 Mbps",
  mbrUL: "108 Mbps",
  mbrDL: "108 Mbps",
  chargingData: {
    chargingMethod: "Online",
    quota: "10000",
    unitCost: "1",
  },
});

export const defaultUpSecurity = (): UpSecurityDTO => ({
  upIntegr: "NOT_NEEDED",
  upConfid: "NOT_NEEDED",
});

export {
  type SubscriptionDTO,
  type FlowRulesDTO,
  type SnssaiConfigurationDTO,
  type DnnConfigurationDTO,
  FlowsMapperImpl,
  SubscriptionMapperImpl,
};
