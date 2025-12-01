import { ReactNode } from "react";
import { Subscription } from "../api";
import { FormProvider, UseFormProps, useForm, useFormContext } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { defaultSubscriptionDTO, subscriptionDTOSchema, type SubscriptionDTO } from "../lib/dtos/subscription";

const SubscriberFormOptions = {
  mode: "onBlur",
  reValidateMode: "onChange",
  resolver: zodResolver(subscriptionDTOSchema),
  defaultValues: defaultSubscriptionDTO(),
} satisfies UseFormProps<Subscription>;

export const SubscriberFormProvider = ({ children }: { children: ReactNode }) => {
  const method = useForm<Subscription>(SubscriberFormOptions);
  return <FormProvider {...method}>{children}</FormProvider>;
};

export const useSubscriptionForm = () => {
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
  } = useFormContext<SubscriptionDTO>();

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
