import { Subscription } from "../../api";
import { validateDNNAMBR, validateMBRGreaterThanGBR, validateSUPIPrefixSameToPLMN } from "./validtors";

export function validateSubscription(subscription: Subscription): { isValid: boolean; error?: string } {
    var validation: { isValid: boolean; error?: string } = { isValid: true };

    validation = validateSUPIPrefixSameToPLMN(subscription);
    if (!validation.isValid) {
        return validation;
    }
    validation = validateDNNAMBR(subscription.SessionManagementSubscriptionData);
    if (!validation.isValid) {
        return validation;
    }
    validation = validateMBRGreaterThanGBR(subscription.QosFlows);
    if (!validation.isValid) {
        return validation;
    }

    return validation;
}