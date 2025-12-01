import { Controller } from "react-hook-form";
import { useSubscriptionForm } from "../../hooks/subscription-form";
import {
  Card,
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

export default function SubscriberFormBasic() {
  const { register, validationErrors, control } = useSubscriptionForm();

  return (
    <Card variant="outlined">
      <Table>
        <TableBody id="Subscriber Data Number">
          <TableRow>
            <TableCell>
              <TextField
                {...register("userNumber", { required: true, valueAsNumber: true })}
                type="number"
                error={validationErrors.userNumber !== undefined}
                helperText={validationErrors.userNumber?.message}
                label="Subscriber data number (auto-incresed with SUPI)"
                variant="outlined"
                required
                fullWidth
              />
            </TableCell>
            <TableCell>
              <TextField
                {...register("ueId", { required: true })}
                error={validationErrors.ueId !== undefined}
                helperText={validationErrors.ueId?.message}
                label="SUPI (IMSI)"
                variant="outlined"
                required
                fullWidth
              />
            </TableCell>
          </TableRow>
        </TableBody>
        <TableBody id="PLMN ID">
          <TableRow>
            <TableCell>
              <TextField
                {...register("plmnID", { required: true })}
                error={validationErrors.plmnID !== undefined}
                helperText={validationErrors.plmnID?.message}
                label="PLMN ID"
                variant="outlined"
                required
                fullWidth
              />
            </TableCell>
            <TableCell>
              <TextField
                {...register("gpsi")}
                error={validationErrors.gpsi !== undefined}
                helperText={validationErrors.gpsi?.message}
                label="GPSI (MSISDN)"
                variant="outlined"
                fullWidth
              />
            </TableCell>
          </TableRow>
        </TableBody>
        <TableBody id="Authentication Management">
          <TableRow>
            <TableCell>
              <TextField
                {...register("auth.authenticationManagementField")}
                error={validationErrors.auth?.authenticationManagementField !== undefined}
                helperText={validationErrors.auth?.authenticationManagementField?.message}
                label="Authentication Management Field (AMF)"
                variant="outlined"
                fullWidth
              />
            </TableCell>
            <TableCell align="left">
              <FormControl variant="outlined" fullWidth>
                <InputLabel>Authentication Method</InputLabel>
                <Controller
                  control={control}
                  name="auth.authenticationMethod"
                  rules={{ required: true }}
                  render={(props) => (
                    <Select
                      {...props.field}
                      error={props.fieldState.error !== undefined}
                      label="Authentication Method"
                      variant="outlined"
                      fullWidth
                      defaultValue=""
                    >
                      <MenuItem value="5G_AKA">5G_AKA</MenuItem>
                      <MenuItem value="EAP_AKA_PRIME">EAP_AKA_PRIME</MenuItem>
                    </Select>
                  )}
                />
              </FormControl>
            </TableCell>
          </TableRow>
        </TableBody>
        <TableBody id="OP">
          <TableRow>
            <TableCell align="left">
              <FormControl variant="outlined" fullWidth>
                <InputLabel>Operator Code Type</InputLabel>
                <Controller
                  control={control}
                  name="auth.operatorCodeType"
                  rules={{ required: true }}
                  render={(props) => (
                    <Select
                      {...props.field}
                      error={props.fieldState.error !== undefined}
                      label="Operator Code Type"
                      variant="outlined"
                      required
                      fullWidth
                      defaultValue=""
                    >
                      <MenuItem value="OP">OP</MenuItem>
                      <MenuItem value="OPc">OPc</MenuItem>
                    </Select>
                  )}
                />
              </FormControl>
            </TableCell>
            <TableCell>
              <TextField
                {...register("auth.operatorCode", {
                  required: true,
                })}
                error={validationErrors.auth?.operatorCode !== undefined}
                helperText={validationErrors.auth?.operatorCode?.message}
                label="Operator Code Value"
                variant="outlined"
                required
                fullWidth
              />
            </TableCell>
          </TableRow>
        </TableBody>
        <TableBody id="SQN">
          <TableRow>
            <TableCell>
              <TextField
                {...register("auth.sequenceNumber", { required: true })}
                error={validationErrors.auth?.sequenceNumber !== undefined}
                helperText={validationErrors.auth?.sequenceNumber?.message}
                label="SQN"
                variant="outlined"
                required
                fullWidth
              />
            </TableCell>
            <TableCell>
              <TextField
                {...register("auth.permanentKey", {
                  required: true,
                })}
                error={validationErrors.auth?.permanentKey !== undefined}
                helperText={validationErrors.auth?.permanentKey?.message}
                label="Permanent Authentication Key"
                variant="outlined"
                required
                fullWidth
              />
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </Card>
  );
}
