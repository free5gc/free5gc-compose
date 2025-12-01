import { Card, Table, TableBody, TableCell, TableRow, TextField } from "@mui/material";
import { useSubscriptionForm } from "../../hooks/subscription-form";

export default function SubscriberFormUeAmbr() {
  const { register, validationErrors } = useSubscriptionForm();

  return (
    <Card variant="outlined">
      <Table>
        <TableBody id="Subscribed UE AMBR">
          <TableRow>
            <TableCell>
              <TextField
                {...register("subscribedUeAmbr.uplink", {
                  required: true,
                })}
                error={validationErrors.subscribedUeAmbr?.uplink !== undefined}
                helperText={validationErrors.subscribedUeAmbr?.uplink?.message}
                label="Uplink"
                variant="outlined"
                required
                fullWidth
              />
            </TableCell>
            <TableCell>
              <TextField
                {...register("subscribedUeAmbr.downlink", {
                  required: true,
                })}
                error={validationErrors.subscribedUeAmbr?.downlink !== undefined}
                helperText={validationErrors.subscribedUeAmbr?.downlink?.message}
                label="Downlink"
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
