import { useProfileForm } from "../../hooks/profile-form";
import { Card, Table, TableBody, TableCell, TableRow, TextField } from "@mui/material";

export default function ProfileFormBasic() {
  const { register, validationErrors } = useProfileForm();

  return (
    <Card variant="outlined">
      <Table>
        <TableBody id="Profile Name">
          <TableRow>
            <TableCell>
              <TextField
                {...register("profileName", { required: true })}
                error={validationErrors.profileName !== undefined}
                helperText={validationErrors.profileName?.message}
                label="Profile Name"
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
