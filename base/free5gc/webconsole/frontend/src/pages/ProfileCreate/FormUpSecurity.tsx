import {
  Button,
  Box,
  Card,
  FormControl,
  Grid,
  InputLabel,
  MenuItem,
  Select,
  Table,
  TableBody,
  TableCell,
  TableRow,
  SelectChangeEvent,
} from "@mui/material";
import { Controller } from "react-hook-form";
import { useProfileForm } from "../../hooks/profile-form";
import { defaultUpSecurity } from "../../lib/dtos/profile";

interface FormUpSecurityProps {
  sessionIndex: number;
  dnnKey: string;
}

function NoUpSecurity(props: FormUpSecurityProps) {
  const { watch, setValue } = useProfileForm();

  const dnnConfig = watch(
    `SnssaiConfigurations.${props.sessionIndex}.dnnConfigurations.${props.dnnKey}`,
  );

  const onUpSecurity = () => {
    dnnConfig.upSecurity = defaultUpSecurity();

    setValue(
      `SnssaiConfigurations.${props.sessionIndex}.dnnConfigurations.${props.dnnKey}`,
      dnnConfig,
    );
  };

  return (
    <div>
      <Table>
        <TableBody>
          <TableRow>
            <TableCell>
              <Button color="secondary" variant="contained" onClick={onUpSecurity} sx={{ m: 0 }}>
                +UP SECURITY
              </Button>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>
  );
}

export default function FormUpSecurity(props: FormUpSecurityProps) {
  const { register, validationErrors, watch, control, getValues, setValue } = useProfileForm();

  const dnnConfig = watch(
    `SnssaiConfigurations.${props.sessionIndex}.dnnConfigurations.${props.dnnKey}`,
  );

  if (!(dnnConfig.upSecurity !== undefined)) {
    return <NoUpSecurity {...props} />;
  }

  const onUpSecurityDelete = () => {
    setValue(`SnssaiConfigurations.${props.sessionIndex}.dnnConfigurations.${props.dnnKey}`, {
      ...dnnConfig,
      upSecurity: undefined,
    });
  };

  return (
    <div>
      <Box sx={{ m: 2 }}>
        <Grid container spacing={2}>
          <Grid item xs={10}>
            <h4>UP Security</h4>
          </Grid>
          <Grid item xs={2}>
            <Box display="flex" justifyContent="flex-end">
              <Button
                color="secondary"
                variant="contained"
                onClick={() => onUpSecurityDelete()}
                sx={{ m: 2, backgroundColor: "red", "&:hover": { backgroundColor: "red" } }}
              >
                DELETE
              </Button>
            </Box>
          </Grid>
        </Grid>

        <Card variant="outlined">
          <Table>
            <TableBody>
              <TableRow>
                <TableCell>
                  <FormControl variant="outlined" fullWidth>
                    <InputLabel>Integrity of UP Security</InputLabel>
                    <Controller
                      control={control}
                      name={`SnssaiConfigurations.${props.sessionIndex}.dnnConfigurations.${props.dnnKey}.upSecurity.upIntegr`}
                      rules={{ required: true }}
                      render={(props) => (
                        <Select
                          {...props.field}
                          error={props.fieldState.error !== undefined}
                          label="Integrity of UP Security"
                          variant="outlined"
                          required
                          fullWidth
                        >
                          <MenuItem value="NOT_NEEDED">NOT_NEEDED</MenuItem>
                          <MenuItem value="PREFERRED">PREFERRED</MenuItem>
                          <MenuItem value="REQUIRED">REQUIRED</MenuItem>
                        </Select>
                      )}
                    />
                  </FormControl>
                </TableCell>
              </TableRow>
            </TableBody>

            <TableBody>
              <TableRow>
                <TableCell>
                  <FormControl variant="outlined" fullWidth>
                    <InputLabel>Confidentiality of UP Security</InputLabel>
                    <Controller
                      control={control}
                      name={`SnssaiConfigurations.${props.sessionIndex}.dnnConfigurations.${props.dnnKey}.upSecurity.upConfid`}
                      rules={{ required: true }}
                      render={(props) => (
                        <Select
                          {...props.field}
                          error={props.fieldState.error !== undefined}
                          label="Confidentiality of UP Security"
                          variant="outlined"
                          required
                          fullWidth
                        >
                          <MenuItem value="NOT_NEEDED">NOT_NEEDED</MenuItem>
                          <MenuItem value="PREFERRED">PREFERRED</MenuItem>
                          <MenuItem value="REQUIRED">REQUIRED</MenuItem>
                        </Select>    
                      )}
                    />
                  </FormControl>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </Card>
      </Box>
    </div>
  );
}
