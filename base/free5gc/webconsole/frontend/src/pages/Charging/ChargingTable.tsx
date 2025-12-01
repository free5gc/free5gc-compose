import React, { useState, useEffect } from "react";
import axios from "../../axios";
import { FlowChargingRecord, ChargingData } from "../../api/api";
import Dashboard from "../../Dashboard";

import ChargingList from "./ChargingList";

import { Button, Grid } from "@mui/material";

export default function ChargingTable() {
  const [expand, setExpand] = useState(true);
  const [refresh, setRefresh] = useState<boolean>(false);
  const [updateTime, setUpdateTime] = useState<Date>(new Date());

  const [onlineChargingData, setOnlineChargingData] = useState<ChargingData[]>([]);
  const [offlineChargingData, setOfflineChargingData] = useState<ChargingData[]>([]);
  const [chargingRecord, setChargingRecord] = useState<FlowChargingRecord[]>([]);

  const fetchChargingRecord = () => {
    const MSG_FETCH_ERROR = "Get Charging Record error";
    axios
      .get("/api/charging-record")
      .then((res) => {
        setChargingRecord(res.data ? res.data : chargingRecord);
        console.log("Charging Record", chargingRecord);
      })
      .catch((err) => {
        console.log(MSG_FETCH_ERROR, err);
      });
  };

  const fetchChargingData = (
    chargingMethod: string,
    setChargingData: React.Dispatch<React.SetStateAction<ChargingData[]>>,
  ) => {
    const MSG_FETCH_ERROR = "Get Charging Data error";
    axios
      .get("/api/charging-data/" + chargingMethod)
      .then((res) => {
        if (res.data) setChargingData(res.data);
      })
      .catch((err) => {
        console.log(MSG_FETCH_ERROR, err);
      });
  };

  const onRefresh = () => {
    fetchChargingData("Online", setOnlineChargingData);
    fetchChargingData("Offline", setOfflineChargingData);
    fetchChargingRecord();
    setUpdateTime(new Date());
  };

  useEffect(() => {
    onRefresh();
  }, [refresh]);

  const onExpand = () => {
    setExpand(!expand);
  };

  return (
    <Dashboard title="UE CHARGING" refreshAction={() => onRefresh()}>
      <Grid container spacing="2">
        <Grid item>
          <Button
            color="secondary"
            variant="contained"
            onClick={() => onExpand()}
            sx={{ m: 2, backgroundColor: "blue", "&:hover": { backgroundColor: "blue" } }}
          >
            {expand ? "Fold" : "Expand"}
          </Button>
        </Grid>
        <Grid item>
          <Button
            color="secondary"
            variant="contained"
            onClick={() => onRefresh()}
            sx={{ m: 2, backgroundColor: "blue", "&:hover": { backgroundColor: "blue" } }}
          >
            Refresh
          </Button>
        </Grid>
        <Grid item>
          <Button color="success" variant="contained" sx={{ m: 2 }} disabled>
            Last update: {updateTime.toISOString().slice(0, 19).replace("T", " ")}
          </Button>
        </Grid>
      </Grid>
      <br />
      <ChargingList
        expand={expand}
        chargingData={offlineChargingData}
        chargingRecord={chargingRecord}
        chargingMethod="Offline"
      />
      <br />
      <ChargingList
        expand={expand}
        chargingData={onlineChargingData}
        chargingRecord={chargingRecord}
        chargingMethod="Online"
      />
    </Dashboard>
  );
}
