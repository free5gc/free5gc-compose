export function toHex(v: number | undefined): string {
  return ("00" + v?.toString(16).toUpperCase()).substr(-2);
}

export function parseDataRate(rate: string | undefined): number {
    if (!rate) return 0;

    const match = rate.match(/^(\d+)\s*(Gbps|Mbps|Kbps|bps)$/i);
    if (!match) return -1;

    const [, value, unit] = match;
    const numValue = parseFloat(value);

    switch (unit.toLowerCase()) {
        case 'gbps':
            return numValue * 1000000;
        case 'mbps':
            return numValue * 1000;
        case 'kbps':
            return numValue;
        case 'bps':
            return numValue / 1000;
        default:
            return 0;
    }
}
