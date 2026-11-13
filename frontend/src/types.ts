export interface SensorData {
  temperature: number;
  humidity: number;
  timestamp: string;
}

export interface RelayState {
  id: number;
  name: string;
  status: boolean;
}

export interface SystemStatus {
  connected: boolean;
  mode: 'NORMAL' | 'TIMEOUT' | 'ERROR';
}