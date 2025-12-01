import {
  Button,
  Box,
  Card,
  Grid,
  Table,
  TableBody,
  TableCell,
  TableRow,
  TextField,
} from "@mui/material";
import type { Nssai } from "../../api";
import { useSubscriptionForm } from "../../hooks/subscription-form";
import { toHex } from "../../lib/utils";
import FormChargingConfig from "./FormCharingConfig";
import { useFieldArray } from "react-hook-form";
import { defaultFlowRule } from "../../lib/dtos/subscription";

interface FormFlowRuleProps {
  snssaiIndex: number;
  dnn: string;
  snssai: Nssai;
}

export default function FormFlowRule({ snssaiIndex, dnn, snssai }: FormFlowRuleProps) {
  const { register, validationErrors, control } = useSubscriptionForm();
  const {
    fields: flowRules,
    append: appendFlowRule,
    remove: removeFlowRule,
  } = useFieldArray({
    control,
    name: `SnssaiConfigurations.${snssaiIndex}.dnnConfigurations.${dnn}.flowRules`,
  });

  const flowKey = toHex(snssai.sst) + snssai.sd;
  const idPrefix = flowKey + "-" + dnn + "-";

  return (
    <>
      {flowRules.map((flow, index) => (
        <div key={flow.id}>
          <Box sx={{ m: 2 }} id={idPrefix + index}>
            <Grid container spacing={2}>
              <Grid item xs={10}>
                <h4>Flow Rules {index + 1}</h4>
              </Grid>
              <Grid item xs={2}>
                <Box display="flex" justifyContent="flex-end">
                  <Button
                    color="secondary"
                    variant="contained"
                    onClick={() => removeFlowRule(index)}
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
                    <TableCell style={{ width: "25%" }}>
                      <TextField
                        {...register(
                          `SnssaiConfigurations.${snssaiIndex}.dnnConfigurations.${dnn}.flowRules.${index}.filter`,
                          {
                            required: true,
                          },
                        )}
                        error={
                          validationErrors.SnssaiConfigurations?.[snssaiIndex]?.dnnConfigurations?.[
                            dnn
                          ]?.flowRules?.[index]?.filter !== undefined
                        }
                        helperText={
                          validationErrors.SnssaiConfigurations?.[snssaiIndex]?.dnnConfigurations?.[
                            dnn
                          ]?.flowRules?.[index]?.filter?.message
                        }
                        label="IP Filter"
                        variant="outlined"
                        required
                        fullWidth
                      />
                    </TableCell>
                    <TableCell style={{ width: "25%" }}>
                      <TextField
                        {...register(
                          `SnssaiConfigurations.${snssaiIndex}.dnnConfigurations.${dnn}.flowRules.${index}.precedence`,
                          {
                            required: true,
                            valueAsNumber: true,
                          },
                        )}
                        error={
                          validationErrors.SnssaiConfigurations?.[snssaiIndex]?.dnnConfigurations?.[
                            dnn
                          ]?.flowRules?.[index]?.precedence !== undefined
                        }
                        helperText={
                          validationErrors.SnssaiConfigurations?.[snssaiIndex]?.dnnConfigurations?.[
                            dnn
                          ]?.flowRules?.[index]?.precedence?.message
                        }
                        label="Precedence"
                        variant="outlined"
                        required
                        fullWidth
                        type="number"
                      />
                    </TableCell>
                    <TableCell style={{ width: "25%" }}>
                      <TextField
                        {...register(
                          `SnssaiConfigurations.${snssaiIndex}.dnnConfigurations.${dnn}.flowRules.${index}.5qi`,
                          {
                            valueAsNumber: true,
                            required: true,
                          },
                        )}
                        error={
                          validationErrors.SnssaiConfigurations?.[snssaiIndex]?.dnnConfigurations?.[
                            dnn
                          ]?.flowRules?.[index]?.["5qi"] !== undefined
                        }
                        helperText={
                          validationErrors.SnssaiConfigurations?.[snssaiIndex]?.dnnConfigurations?.[
                            dnn
                          ]?.flowRules?.[index]?.["5qi"]?.message
                        }
                        label="5QI"
                        variant="outlined"
                        required
                        fullWidth
                        type="number"
                      />
                    </TableCell>

                    <TableCell style={{ width: "25%" }}>{/* Keep layout aligned*/}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell style={{ width: "25%" }}>
                      <TextField
                        {...register(
                          `SnssaiConfigurations.${snssaiIndex}.dnnConfigurations.${dnn}.flowRules.${index}.gbrUL`,
                          {
                            required: true,
                          },
                        )}
                        error={
                          validationErrors.SnssaiConfigurations?.[snssaiIndex]?.dnnConfigurations?.[
                            dnn
                          ]?.flowRules?.[index]?.gbrUL !== undefined
                        }
                        helperText={
                          validationErrors.SnssaiConfigurations?.[snssaiIndex]?.dnnConfigurations?.[
                            dnn
                          ]?.flowRules?.[index]?.gbrUL?.message
                        }
                        label="Uplink GBR"
                        variant="outlined"
                        required
                        fullWidth
                      />
                    </TableCell>
                    <TableCell style={{ width: "25%" }}>
                      <TextField
                        {...register(
                          `SnssaiConfigurations.${snssaiIndex}.dnnConfigurations.${dnn}.flowRules.${index}.gbrDL`,
                          {
                            required: true,
                          },
                        )}
                        error={
                          validationErrors.SnssaiConfigurations?.[snssaiIndex]?.dnnConfigurations?.[
                            dnn
                          ]?.flowRules?.[index]?.gbrDL !== undefined
                        }
                        helperText={
                          validationErrors.SnssaiConfigurations?.[snssaiIndex]?.dnnConfigurations?.[
                            dnn
                          ]?.flowRules?.[index]?.gbrDL?.message
                        }
                        label="Downlink GBR"
                        variant="outlined"
                        required
                        fullWidth
                      />
                    </TableCell>
                    <TableCell style={{ width: "25%" }}>
                      <TextField
                        {...register(
                          `SnssaiConfigurations.${snssaiIndex}.dnnConfigurations.${dnn}.flowRules.${index}.mbrUL`,
                          {
                            required: true,
                          },
                        )}
                        error={
                          validationErrors.SnssaiConfigurations?.[snssaiIndex]?.dnnConfigurations?.[
                            dnn
                          ]?.flowRules?.[index]?.mbrUL !== undefined
                        }
                        helperText={
                          validationErrors.SnssaiConfigurations?.[snssaiIndex]?.dnnConfigurations?.[
                            dnn
                          ]?.flowRules?.[index]?.mbrUL?.message
                        }
                        label="Uplink MBR"
                        variant="outlined"
                        required
                        fullWidth
                      />
                    </TableCell>
                    <TableCell style={{ width: "25%" }}>
                      <TextField
                        {...register(
                          `SnssaiConfigurations.${snssaiIndex}.dnnConfigurations.${dnn}.flowRules.${index}.mbrDL`,
                          {
                            required: true,
                          },
                        )}
                        error={
                          validationErrors.SnssaiConfigurations?.[snssaiIndex]?.dnnConfigurations?.[
                            dnn
                          ]?.flowRules?.[index]?.mbrDL !== undefined
                        }
                        helperText={
                          validationErrors.SnssaiConfigurations?.[snssaiIndex]?.dnnConfigurations?.[
                            dnn
                          ]?.flowRules?.[index]?.mbrDL?.message
                        }
                        label="Downlink MBR"
                        variant="outlined"
                        required
                        fullWidth
                      />
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
              <FormChargingConfig snssaiIndex={snssaiIndex} dnn={dnn} filterIndex={index} />
            </Card>
          </Box>
        </div>
      ))}
      <Table>
        <TableBody>
          <TableRow>
            <TableCell>
              <Button
                color="secondary"
                variant="outlined"
                onClick={() => appendFlowRule(defaultFlowRule())}
                sx={{ m: 0 }}
              >
                +FLOW RULE
              </Button>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </>
  );
}
