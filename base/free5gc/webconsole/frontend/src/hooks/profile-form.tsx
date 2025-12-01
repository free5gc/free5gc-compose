import { ReactNode } from "react";
import { Profile } from "../api";
import { FormProvider, UseFormProps, useForm, useFormContext } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { defaultProfileDTO, profileDTOSchema, type ProfileDTO } from "../lib/dtos/profile";

const ProfileFormOptions = {
    mode: "onBlur",
    reValidateMode: "onChange",
    resolver: zodResolver(profileDTOSchema),
    defaultValues: defaultProfileDTO(),
} satisfies UseFormProps<Profile>;

export const ProfileFormProvider = ({ children }: { children: ReactNode }) => {
    const method = useForm<Profile>(ProfileFormOptions);
    return <FormProvider {...method}>{children}</FormProvider>;
};  

export const useProfileForm = () => {
    const {
        register,
        handleSubmit,
        watch,
        getValues,
        setValue,
        setFocus,
        reset,
        control,
        formState: { errors: validationErrors },
    } = useFormContext<ProfileDTO>();

    return {
        register,
        validationErrors,
        handleSubmit,
        watch,
        getValues,
        control,
        setValue,
        setFocus,
        reset,
    };
};
