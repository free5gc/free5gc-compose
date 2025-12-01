import { test, describe } from "vitest";
import { Profile } from "../../api";
import {
    FlowsMapperImpl,
    ProfileMapperImpl,
    ProfileDTO,
    defaultProfileDTO,
    defaultDnnConfig,
    defaultUpSecurity,
    upSecurityDTOSchema,
    dnnConfigurationDTOSchema,
} from "./profile";
import assert from "node:assert";

const defaultProfile = (): Profile => ({
    profileName: "profile-1",
    AccessAndMobilitySubscriptionData: {
        subscribedUeAmbr: {
            uplink: "1 Gbps",
            downlink: "2 Gbps",
        },
        nssai: {
            defaultSingleNssais: [
                {
                    sst: 1,
                    sd: "010203",
                },
            ],
            singleNssais: [
                {
                    sst: 1,
                    sd: "112233",
                },
            ],
        },
    },
    SessionManagementSubscriptionData: [
        {
            singleNssai: {
                sst: 1,
                sd: "010203",
            },
            dnnConfigurations: {
                internet: {
                    pduSessionTypes: {
                        defaultSessionType: "IPV4",
                        allowedSessionTypes: ["IPV4"],
                    },
                    sscModes: {
                        defaultSscMode: "SSC_MODE_1",
                        allowedSscModes: ["SSC_MODE_2", "SSC_MODE_3"],
                    },
                    "5gQosProfile": {
                        "5qi": 9,
                        arp: {
                            priorityLevel: 8,
                            preemptCap: "",
                            preemptVuln: "",
                        },
                        priorityLevel: 8,
                    },
                    sessionAmbr: {
                        uplink: "1000 Mbps",
                        downlink: "1000 Mbps",
                    },
                    staticIpAddress: [],
                },
            },
        },
        {
            singleNssai: {
                sst: 1,
                sd: "112233",
            },
            dnnConfigurations: {
                internet: {
                    pduSessionTypes: {
                        defaultSessionType: "IPV4",
                        allowedSessionTypes: ["IPV4"],
                    },
                    sscModes: {
                        defaultSscMode: "SSC_MODE_1",
                        allowedSscModes: ["SSC_MODE_2", "SSC_MODE_3"],
                    },
                    "5gQosProfile": {
                        "5qi": 8,
                        arp: {
                            priorityLevel: 8,
                            preemptCap: "",
                            preemptVuln: "",
                        },
                        priorityLevel: 8,
                    },
                    sessionAmbr: {
                        uplink: "1000 Mbps",
                        downlink: "1000 Mbps",
                    },
                    staticIpAddress: [],
                },
            },
        },
    ],
    SmfSelectionSubscriptionData: {
        subscribedSnssaiInfos: {
            "01010203": {
                dnnInfos: [
                    {
                        dnn: "internet",
                    },
                ],
            },
            "01112233": {
                dnnInfos: [
                    {
                        dnn: "internet",
                    },
                ],
            },
        },
    },
    AmPolicyData: {
        subscCats: ["free5gc"],
    },
    SmPolicyData: {
        smPolicySnssaiData: {
            "01010203": {
                snssai: {
                    sst: 1,
                    sd: "010203",
                },
                smPolicyDnnData: {
                    internet: {
                        dnn: "internet",
                    },
                },
            },
            "01112233": {
                snssai: {
                    sst: 1,
                    sd: "112233",
                },
                smPolicyDnnData: {
                    internet: {
                        dnn: "internet",
                    },
                },
            },
        },
    },
    FlowRules: [
        {
            filter: "1.1.1.1/32",
            precedence: 128,
            snssai: "01010203",
            dnn: "internet",
            qosRef: 1,
        },
        {
            filter: "1.1.1.1/32",
            precedence: 127,
            snssai: "01112233",
            dnn: "internet",
            qosRef: 2,
        },
    ],
    QosFlows: [
        {
            snssai: "01010203",
            dnn: "internet",
            qosRef: 1,
            "5qi": 8,
            mbrUL: "208 Mbps",
            mbrDL: "208 Mbps",
            gbrUL: "108 Mbps",
            gbrDL: "108 Mbps",
        },
        {
            snssai: "01112233",
            dnn: "internet",
            qosRef: 2,
            "5qi": 7,
            mbrUL: "407 Mbps",
            mbrDL: "407 Mbps",
            gbrUL: "207 Mbps",
            gbrDL: "207 Mbps",
        },
    ],
    ChargingDatas: [
        {
            snssai: "01010203",
            dnn: "",
            filter: "",
            chargingMethod: "Offline",
            quota: "100000",
            unitCost: "1",
        },
        {
            snssai: "01010203",
            dnn: "internet",
            qosRef: 1,
            filter: "1.1.1.1/32",
            chargingMethod: "Offline",
            quota: "100000",
            unitCost: "1",
        },
        {
            snssai: "01112233",
            dnn: "",
            filter: "",
            chargingMethod: "Online",
            quota: "100000",
            unitCost: "1",
        },
        {
            snssai: "01112233",
            dnn: "internet",
            qosRef: 2,
            filter: "1.1.1.1/32",
            chargingMethod: "Online",
            quota: "5000",
            unitCost: "1",
        },
    ],
});

describe("ProfileDTO", () => {
    describe("default profile", () => {
        test("build", ({ expect }) => {
            const profileBuilder = new ProfileMapperImpl(new FlowsMapperImpl());
            const profile = profileBuilder.mapFromDto(defaultProfileDTO());
            assert.deepEqual(JSON.parse(JSON.stringify(profile)), defaultProfile());
        });

        test("parse", ({ expect }) => {
            const profileBuilder = new ProfileMapperImpl(new FlowsMapperImpl());
            const profile = profileBuilder.mapFromProfile(defaultProfile());
            assert.deepEqual(JSON.parse(JSON.stringify(profile)), defaultProfileDTO());
        });
    });

    describe("profile with up security", () => {
        const profileWithUpSecurity: Profile = defaultProfile();
        profileWithUpSecurity.SessionManagementSubscriptionData[0].dnnConfigurations!["internet"].upSecurity = {
            upIntegr: "1 Gbps",
            upConfid: "2 Gbps",
        };

        const profileWithUpSecurityDTO: ProfileDTO = defaultProfileDTO();
        profileWithUpSecurityDTO.SnssaiConfigurations[0].dnnConfigurations["internet"].upSecurity = {
            upIntegr: "1 Gbps",
            upConfid: "2 Gbps",
        };

        test("build", ({ expect }) => {
            const profileBuilder = new ProfileMapperImpl(new FlowsMapperImpl());
            const profile = profileBuilder.mapFromDto(profileWithUpSecurityDTO);
            assert.deepEqual(JSON.parse(JSON.stringify(profile)), profileWithUpSecurity);
        });

        test("parse", () => {
            const profileBuilder = new ProfileMapperImpl(new FlowsMapperImpl());
            const profile = profileBuilder.mapFromProfile(profileWithUpSecurity);
            assert.deepEqual(
                JSON.parse(JSON.stringify(profile)),
                JSON.parse(JSON.stringify(profileWithUpSecurityDTO)),
            );
        });
    });

    describe("default values", () => {
        test("defaultDnnConfig should match schema", ({ expect }) => {
            const result = defaultDnnConfig();
            expect(() => {
                dnnConfigurationDTOSchema.parse(result);
            }).not.toThrow();
        });

        test("defaultUpSecurity should match schema", ({ expect }) => {
            const result = defaultUpSecurity();
            expect(() => {
                upSecurityDTOSchema.parse(result);
            }).not.toThrow();
        });
    });
});
