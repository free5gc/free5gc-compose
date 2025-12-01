import { z } from "zod";
import {
    FlowRules,
    QosFlows,
    ChargingData,
    SubscribedUeAmbr,
    Nssai,
    SessionManagementSubscriptionData,
    DnnConfiguration,
    UpSecurity,
    Profile,
} from "../../api"
import { DEFAULT_5QI } from "../const";

interface AmbrDTO {
    uplink: string;
    downlink: string;
}

export const ambrDTOSchema = z.object({
    uplink: z.string(),
    downlink: z.string(),
})

interface ChargingDataDTO {
    chargingMethod: "Online" | "Offline";
    quota: string;
    unitCost: string;
}

export const chargingDataDTOSchema = z.object({
    chargingMethod: z.enum(["Online", "Offline"]),
    quota: z.string(),
    unitCost: z.string(),
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

interface UpSecurityDTO {
    upIntegr: string;
    upConfid: string;
}

export const upSecurityDTOSchema = z.object({
    upIntegr: z.string(),
    upConfid: z.string(),
})

interface DnnConfigurationDTO {
    default5qi: number;
    sessionAmbr: AmbrDTO;
    enableStaticIpv4Address: boolean;
    staticIpv4Address?: string;
    flowRules: FlowRulesDTO[];
    upSecurity?: UpSecurityDTO;
}

export const dnnConfigurationDTOSchema = z.object({
    default5qi: z.number(),
    sessionAmbr: ambrDTOSchema,
    enableStaticIpv4Address: z.boolean(),
    staticIpv4Address: z.string().optional(),
    flowRules: z.array(flowRulesDTOSchema),
    upSecurity: upSecurityDTOSchema.optional(),
})

interface SnssaiConfigurationDTO {
    sst: number;
    sd: string;
    isDefault: boolean;
    chargingData: ChargingDataDTO;
    dnnConfigurations: { [key: string]: DnnConfigurationDTO };
}

export const snssaiConfigurationDTOSchema = z.object({
    sst: z.number(),
    sd: z.string(),
    isDefault: z.boolean(),
    chargingData: chargingDataDTOSchema,
    dnnConfigurations: z.record(dnnConfigurationDTOSchema),
})

interface ProfileDTO {
    profileName: string;
    subscribedUeAmbr: AmbrDTO;
    SnssaiConfigurations: SnssaiConfigurationDTO[];
}

export const profileDTOSchema = z.object({
    profileName: z.string(),
    subscribedUeAmbr: ambrDTOSchema,
    SnssaiConfigurations: z.array(snssaiConfigurationDTOSchema),
})

interface FlowsDTO {
    flowRules: FlowRules[];
    qosFlows: QosFlows[];
    chargingDatas: ChargingData[];
}

interface FlowsMapper {
    map(profile: ProfileDTO): FlowsDTO;
}

class FlowsMapperImpl implements FlowsMapper {
    refNumber: number = 1;
    flowRules: FlowRules[] = [];
    qosFlows: QosFlows[] = [];
    chargingDatas: ChargingData[] = [];

    private buildDnns(profile: ProfileDTO): {
        snssai: string;
        dnn: string;
        sliceCharingData: ChargingDataDTO;
        flowRules: FlowRulesDTO[];
    }[] {
        return profile.SnssaiConfigurations.reduce(
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

    map(profile: ProfileDTO): FlowsDTO {
        const dnns = this.buildDnns(profile);

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

interface ProfileMapper {
    mapFromDto(profile: ProfileDTO): Profile;
    mapFromProfile(profile: Profile): ProfileDTO;
}

class ProfileMapperImpl implements ProfileMapper {
    constructor(private readonly flowsBuilder: FlowsMapper) {}

    mapFromDto(profile: ProfileDTO): Profile {
        const flows = this.flowsBuilder.map(profile);

        return {
            profileName: profile.profileName,
            AccessAndMobilitySubscriptionData: {
                subscribedUeAmbr: this.buildSubscriberAmbr(profile.subscribedUeAmbr),
                nssai: {
                    defaultSingleNssais: profile.SnssaiConfigurations.filter((s) => s.isDefault).map(
                        (s) => this.buildNssai(s)),
                    singleNssais: profile.SnssaiConfigurations.filter((s) => !s.isDefault).map(
                        (s) => this.buildNssai(s)),
                }
            },
            SessionManagementSubscriptionData: profile.SnssaiConfigurations.map((s) =>
                this.buildSessionManagementSubscriptionData(s),
            ),

            SmfSelectionSubscriptionData: {
                subscribedSnssaiInfos: Object.fromEntries(
                    profile.SnssaiConfigurations.map((s) => [
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
                    profile.SnssaiConfigurations.map((s) => [
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
        }
    }

    mapFromProfile(profile: Profile): ProfileDTO {
        return {
            profileName: profile.profileName,
            subscribedUeAmbr: {
                uplink: profile.AccessAndMobilitySubscriptionData.subscribedUeAmbr?.uplink ?? "",
                downlink: profile.AccessAndMobilitySubscriptionData.subscribedUeAmbr?.downlink ?? "",
            },
            SnssaiConfigurations: profile.SessionManagementSubscriptionData.map((s) => ({
                sst: s.singleNssai.sst,
                sd: s.singleNssai.sd ?? "",
                isDefault: this.snssaiIsDefault(s.singleNssai, profile),
                chargingData: this.findSliceChargingData(s.singleNssai, profile),
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
                            flowRules: this.parseDnnFlowRules(s.singleNssai, key, profile),
                            upSecurity: value.upSecurity,
                        } satisfies DnnConfigurationDTO,
                    ]),
                ),
            })),
        }
    }

    private snssaiIsDefault(nssai: Nssai, profile: Profile): boolean {
        return (
            profile.AccessAndMobilitySubscriptionData.nssai?.defaultSingleNssais?.some(
                (n) => n.sst === nssai.sst && n.sd === nssai.sd,
            ) ?? false
        );
    }

    private findSliceChargingData(nssai: Nssai, profile: Profile): ChargingDataDTO {
        const chargingData = profile.ChargingDatas.find((c) => {
            if (c.dnn !== "" || c.filter !== "") {
                return false;
            }

            return c.snssai === nssai.sst.toString().padStart(2, "0") + nssai.sd;
        });

        return {
            chargingMethod: chargingData?.chargingMethod === "Online" ? "Online" : "Offline",
            quota: chargingData?.quota ?? "",
            unitCost: chargingData?.unitCost ?? "",
        };
    }

    private parseDnnFlowRules(snssai: Nssai, dnn: string, profile: Profile): FlowRulesDTO[] {
        const qosFlows = profile.QosFlows.filter(
            (f) => f.dnn === dnn && f.snssai === snssai.sst.toString().padStart(2, "0") + snssai.sd,
        );

        return qosFlows.map((f) => {
            const flowRule = profile.FlowRules.find((r) => r.qosRef === f.qosRef);
            const chargingData = profile.ChargingDatas.find((c) => c.qosRef === f.qosRef);
        
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
        })
    }

    buildSubscriberAmbr(data: AmbrDTO): SubscribedUeAmbr{
        return {
            uplink: data.uplink,
            downlink: data.downlink,
        }
    }

    buildNssai(data: SnssaiConfigurationDTO): Nssai {
        return {
            sst: data.sst,
            sd: data.sd,
        }
    }

    buildSessionManagementSubscriptionData(data: SnssaiConfigurationDTO): SessionManagementSubscriptionData {
        return {
            singleNssai: this.buildNssai(data),
            dnnConfigurations: Object.fromEntries(
                Object.entries(data.dnnConfigurations).map(([key, value]) => [
                    key,
                    this.buildDnnConfiguration(value),
                ]),
            ),
        }
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
        }
    }

    buildUpSecurity(data: UpSecurityDTO | undefined): UpSecurity | undefined {
        if (!data) {
            return undefined;
        }

        return {
            upIntegr: data.upIntegr,
            upConfid: data.upConfid,
        }
    }
}

export const defaultProfileDTO = (): ProfileDTO => ({
    profileName: "profile-1",
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
})

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
    type ProfileDTO,
    type FlowsDTO,
    type SnssaiConfigurationDTO,
    type DnnConfigurationDTO,
    ProfileMapperImpl,
    FlowsMapperImpl,
}
