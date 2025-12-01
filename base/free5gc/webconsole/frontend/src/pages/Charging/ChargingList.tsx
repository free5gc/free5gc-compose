import React from "react";
import { Table, TableBody, TableCell, TableHead, TableRow } from "@mui/material";
import { ChargingData, FlowChargingRecord } from "../../api/api";

const ChargingList: React.FC<{
  expand: boolean;
  chargingData: ChargingData[];
  chargingRecord: FlowChargingRecord[];
  chargingMethod: string;
}> = (props) => {
  const tableColumnNames = [
    "SUPI",
    "S-NSSAI",
    "DNN",
    "IP Filter",
    props.chargingMethod === "Online" ? "Quota" : "Usage",
    "Data Total Volume",
    "Data Volume UL",
    "Data Volume DL",
  ];

  const FlowUsageCell: React.FC<{
    supi: string;
    dnn: string;
    snssai: string;
    filter: string;
  }> = (Props) => {
    const chargingRecordMatch = props.chargingRecord.find(
      (a) =>
        a.Supi === Props.supi! &&
        a.Dnn! === Props.dnn &&
        a.Snssai! === Props.snssai &&
        a.Filter! === Props.filter,
    );

    return (
      <>
        <TableCell>
          {chargingRecordMatch
            ? props.chargingMethod === "Online"
              ? chargingRecordMatch.QuotaLeft
              : chargingRecordMatch.Usage
            : "-"}
        </TableCell>
        <TableCell>{chargingRecordMatch ? chargingRecordMatch.TotalVol : "-"}</TableCell>
        <TableCell>{chargingRecordMatch ? chargingRecordMatch.UlVol : "-"}</TableCell>
        <TableCell>{chargingRecordMatch ? chargingRecordMatch.DlVol : "-"}</TableCell>
      </>
    );
  };

  const PerFlowTableView: React.FC<{ Supi: string; Snssai: string }> = (Props) => {
    if (!props.expand) return <></>;

    return (
      <>
        {props.chargingData
          .filter((a) => a!.filter !== "" && a!.ueId === Props.Supi && a!.snssai === Props.Snssai)
          .map((cd, idx) => (
            <TableRow key={idx}>
              <TableCell></TableCell>
              <TableCell>{cd.snssai}</TableCell>
              <TableCell>{cd.dnn}</TableCell>
              <TableCell>{cd.filter}</TableCell>
              {
                <FlowUsageCell
                  supi={Props.Supi}
                  dnn={cd.dnn!}
                  snssai={Props.Snssai}
                  filter={cd.filter!}
                />
              }
            </TableRow>
          ))}
      </>
    );
  };

  return (
    <Table>
      <TableHead>
        <TableRow>
          <TableCell colSpan={tableColumnNames.length}>
            <h3>{props.chargingMethod} Charging</h3>
          </TableCell>
        </TableRow>
        <TableRow>
          {tableColumnNames.map((colName, idx) => {
            return <TableCell key={idx}>{colName}</TableCell>;
          })}
        </TableRow>
      </TableHead>
      <TableBody>
        {props.chargingData
          .filter((a) => a!.filter === "" && a!.dnn === "")
          .map((cd, idx) => {
            return (
              <React.Fragment key={idx}>
                <TableRow>
                  <TableCell>{cd.ueId}</TableCell>
                  <TableCell>{cd.snssai}</TableCell>
                  <TableCell>{cd.dnn}</TableCell>
                  <TableCell>{cd.filter}</TableCell>
                  <FlowUsageCell
                    supi={cd.ueId!}
                    dnn={cd.dnn!}
                    snssai={cd.snssai!}
                    filter={cd.filter!}
                  />
                </TableRow>
                {<PerFlowTableView Supi={cd.ueId!} Snssai={cd.snssai!} />}
              </React.Fragment>
            );
          })}
      </TableBody>
    </Table>
  );
};

export default ChargingList;
