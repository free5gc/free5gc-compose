import { SessionManagementSubscriptionData, Subscription } from "../../api";
import { parseDataRate } from "../utils";

export function validateSUPIPrefixSameToPLMN(subscription: Subscription): { isValid: boolean; error?: string } {
    const supi = subscription.ueId;
    const plmn = subscription.plmnID;
    const supiPrefix = supi.substring(5, 10);
    if (supiPrefix !== plmn) {
        return { isValid: false, error: "SUPI Prefix must be same as PLMN" };
    }
    return { isValid: true };
}

export function validateDNNAMBR(sessionManagementSubscriptionDatas: SessionManagementSubscriptionData[]): { isValid: boolean; error?: string } {
    for (let i = 0; i < sessionManagementSubscriptionDatas.length; i++) {
        const sessionManagementSubscriptionData = sessionManagementSubscriptionDatas[i];
        if (!sessionManagementSubscriptionData.dnnConfigurations) {
            return { isValid: true };
        }

        for (const dnn in sessionManagementSubscriptionData.dnnConfigurations) {
            const dnnConfiguration = sessionManagementSubscriptionData.dnnConfigurations[dnn];
            if (!dnnConfiguration.sessionAmbr) {
                return { isValid: true };
            }

            const uplinkAmbr = dnnConfiguration.sessionAmbr.uplink;
            const downlinkAmbr = dnnConfiguration.sessionAmbr.downlink;
            const uplinkFlow = parseDataRate(uplinkAmbr);

            if (uplinkFlow === -1) {
                return {
                    isValid: false,
                    error: "In S-NSSAI " + sessionManagementSubscriptionData.singleNssai.sd + "'s DNN: " + dnn + "\nuplink AMBR is invalid"
                };
            }

            const downlinkFlow = parseDataRate(downlinkAmbr);
            if (downlinkFlow === -1) {
                return {
                    isValid: false,
                    error: "In S-NSSAI " + sessionManagementSubscriptionData.singleNssai.sd + "'s DNN: " + dnn + "\ndownlink AMBR is invalid"
                };
            }
        }
    }

    return { isValid: true };
}

export function validateMBRGreaterThanGBR(QosFlows: any[]): { isValid: boolean; error?: string } {
    for (let i = 0; i < QosFlows.length; i++) {
        const qosFlow = QosFlows[i];
        const gbrDL = parseDataRate(qosFlow.gbrDL);
        if (gbrDL === -1) {
            return {
                isValid: false,
                error: `In S-NSSAI ${qosFlow.snssai}'s Flow Rule\nDownlink GBR is invalid`
            };
        }

        const mbrDL = parseDataRate(qosFlow.mbrDL);
        if (mbrDL === -1) {
            return {
                isValid: false,
                error: `In S-NSSAI ${qosFlow.snssai}'s Flow Rule\nDownlink MBR is invalid`
            };
        }

        const gbrUL = parseDataRate(qosFlow.gbrUL);
        if (gbrUL === -1) {
            return {
                isValid: false,
                error: `In S-NSSAI ${qosFlow.snssai}'s Flow Rule\nUplink GBR is invalid`
            };
        }

        const mbrUL = parseDataRate(qosFlow.mbrUL);
        if (mbrUL === -1) {
            return {
                isValid: false,
                error: `In S-NSSAI ${qosFlow.snssai}'s Flow Rule\nUplink MBR is invalid`
            };
        }
    }

    return { isValid: true };
}