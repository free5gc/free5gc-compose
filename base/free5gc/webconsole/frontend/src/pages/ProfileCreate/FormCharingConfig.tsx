import {
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  Table,
  TableBody,
  TableCell,
  TableRow,
  TextField,
} from "@mui/material";
import { useProfileForm } from "../../hooks/profile-form";
import { Controller } from "react-hook-form";

interface FormCharginConfigProps {
  snssaiIndex: number;
  dnn?: string;
  filterIndex?: number;
}

function FormSliceChargingConfig({ snssaiIndex }: FormCharginConfigProps) {
  const { register, validationErrors, watch, control } = useProfileForm();

  const isOnlineCharging =
    watch(`SnssaiConfigurations.${snssaiIndex}.chargingData.chargingMethod`) === "Online";

  return (
    <Table>
      <TableBody id={snssaiIndex + "-charging-config"}>
        <TableRow>
          <TableCell style={{ width: "33%" }}>
            <FormControl variant="outlined" fullWidth>
              <InputLabel>Charging Method</InputLabel>
              <Controller
                control={control}
                name={`SnssaiConfigurations.${snssaiIndex}.chargingData.chargingMethod`}
                rules={{ required: true }}
                render={(props) => (
                  <Select
                    {...props.field}
                    error={props.fieldState.error !== undefined}
                    label="Charging Method"
                    variant="outlined"
                    required
                    fullWidth
                    defaultValue=""
                  >
                    <MenuItem value="Offline">Offline</MenuItem>
                    <MenuItem value="Online">Online</MenuItem>
                  </Select>
                )}
              />
            </FormControl>
          </TableCell>

          <TableCell style={{ width: "33%" }}>
            <TextField
              {...register(`SnssaiConfigurations.${snssaiIndex}.chargingData.quota`, {
                required: true,
              })}
              error={
                validationErrors.SnssaiConfigurations?.[snssaiIndex]?.chargingData?.quota !==
                undefined
              }
              helperText={
                validationErrors.SnssaiConfigurations?.[snssaiIndex]?.chargingData?.quota?.message
              }
              label="Quota (monetary)"
              variant="outlined"
              required={isOnlineCharging}
              disabled={!isOnlineCharging}
              fullWidth
            />
          </TableCell>

          <TableCell style={{ width: "33%" }}>
            <TextField
              {...register(`SnssaiConfigurations.${snssaiIndex}.chargingData.unitCost`, {
                required: true,
              })}
              error={
                validationErrors.SnssaiConfigurations?.[snssaiIndex]?.chargingData?.unitCost !==
                undefined
              }
              helperText={
                validationErrors.SnssaiConfigurations?.[snssaiIndex]?.chargingData?.unitCost
                  ?.message
              }
              label="Unit Cost (money per byte)"
              variant="outlined"
              required
              fullWidth
            />
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
  );
}

function FormFlowChargingConfig({ snssaiIndex, dnn, filterIndex }: FormCharginConfigProps) {
  const { register, validationErrors, watch, control } = useProfileForm();

  if (dnn === undefined) {
    throw new Error("dnn is undefined");
  }
  if (filterIndex === undefined) {
    throw new Error("filterIndex is undefined");
  }

  const isOnlineCharging =
    watch(
      `SnssaiConfigurations.${snssaiIndex}.dnnConfigurations.${dnn}.flowRules.${filterIndex}.chargingData.chargingMethod`,
    ) === "Online";

  return (
    <Table>
      <TableBody id={`${snssaiIndex}-${dnn}-${filterIndex}-charging-config`}>
        <TableRow>
          <TableCell style={{ width: "33%" }}>
            <FormControl variant="outlined" fullWidth>
              <InputLabel>Charging Method</InputLabel>
              <Controller
                control={control}
                name={`SnssaiConfigurations.${snssaiIndex}.dnnConfigurations.${dnn}.flowRules.${filterIndex}.chargingData.chargingMethod`}
                rules={{ required: true }}
                render={(props) => (
                  <Select
                    {...props.field}
                    error={props.fieldState.error !== undefined}
                    label="Charging Method"
                    variant="outlined"
                    required
                    fullWidth
                    defaultValue=""
                  >
                    <MenuItem value="Offline">Offline</MenuItem>
                    <MenuItem value="Online">Online</MenuItem>
                  </Select>
                )}
              />
            </FormControl>
          </TableCell>
          <TableCell style={{ width: "33%" }}>
            <TextField
              {...register(
                `SnssaiConfigurations.${snssaiIndex}.dnnConfigurations.${dnn}.flowRules.${filterIndex}.chargingData.quota`,
                {
                  required: true,
                },
              )}
              error={
                validationErrors.SnssaiConfigurations?.[snssaiIndex]?.dnnConfigurations?.[dnn]
                  ?.flowRules?.[filterIndex]?.chargingData?.quota !== undefined
              }
              helperText={
                validationErrors.SnssaiConfigurations?.[snssaiIndex]?.dnnConfigurations?.[dnn]
                  ?.flowRules?.[filterIndex]?.chargingData?.quota?.message
              }
              label="Quota (monetary)"
              variant="outlined"
              required={isOnlineCharging}
              disabled={!isOnlineCharging}
              fullWidth
            />
          </TableCell>
          <TableCell style={{ width: "33%" }}>
            <TextField
              {...register(
                `SnssaiConfigurations.${snssaiIndex}.dnnConfigurations.${dnn}.flowRules.${filterIndex}.chargingData.unitCost`,
                {
                  required: true,
                },
              )}
              error={
                validationErrors.SnssaiConfigurations?.[snssaiIndex]?.dnnConfigurations?.[dnn]
                  ?.flowRules?.[filterIndex]?.chargingData?.unitCost !== undefined
              }
              helperText={
                validationErrors.SnssaiConfigurations?.[snssaiIndex]?.dnnConfigurations?.[dnn]
                  ?.flowRules?.[filterIndex]?.chargingData?.unitCost?.message
              }
              label="Unit Cost (money per byte)"
              variant="outlined"
              required
              fullWidth
            />
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
  );
}

export default function FormChargingConfig(props: FormCharginConfigProps) {
  if (props.dnn === undefined) {
    return <FormSliceChargingConfig {...props} />;
  }

  return <FormFlowChargingConfig {...props} />;
}
