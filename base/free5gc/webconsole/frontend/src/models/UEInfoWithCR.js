export default class UEInfoWithCR {
    supi = "";
    status = "";
    totalVol = 0;
    ulVol = 0;
    dlVol = 0;
    quotaLeft = 0;
    flowInfos = []
  
    constructor(supi, status, totalVol = 0, ulVol = 0, dlVol = 0, quotaLeft = 0, flowInfos = []) {
      this.supi = supi;
      this.status = status;
      this.totalVol = totalVol;
      this.ulVol = ulVol;
      this.dlVol = dlVol;
      this.quotaLeft = quotaLeft;
      this.flowInfos = flowInfos;
    }
  }
  