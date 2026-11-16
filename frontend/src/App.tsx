import React, { useState, useEffect } from 'react';
import { useAuth } from './context/AuthContext';

interface SensorConfig {
  name: string;
  address: number;
  unit: string;
}

interface RelayConfig {
  id: number;
  name: string;
  coil_address: number;
}

interface DeviceConfig {
  id: number;
  name: string;
  url: string;
  sensors: SensorConfig[];
  relays: RelayConfig[];
}

export function App() {
  const { token, user, logout, hasPermission, login } = useAuth();
  const [usernameInput, setUsernameInput] = useState('');
  const [passwordInput, setPasswordInput] = useState('');
  const [loginError, setLoginError] = useState('');
  
  const [sensorsMap, setSensorsMap] = useState<{ [key: number]: { [address: number]: number } }>({});
  const [devices, setDevices] = useState<DeviceConfig[]>([]);
  const [relayStates, setRelayStates] = useState<{ [key: string]: boolean }>("{}" as any);
  const [logMessage, setLogMessage] = useState<string>('Sistem başlatıldı. Backend bekleniyor...');

  useEffect(() => {
    if (!token) return;

    const headers = {
      'Authorization': 'Bearer ' + token,
      'Content-Type': 'application/json'
    };

    // 1. Cihazları ve içindeki config'leri alma (/api/devices)
    fetch('http://localhost:8080/api/devices', { headers })
      .then(res => {
        if (!res.ok) throw new Error('Cihazlar alınamadı.');
        return res.json();
      })
      .then((data: DeviceConfig[]) => {
        setDevices(data);
      })
      .catch(err => setLogMessage(`Uyarı: ${err.message}`));

    // 2. Periyodik Sensör Polling (/api/sensors)
    const fetchSensors = async () => {
      try {
        const response = await fetch('http://localhost:8080/api/sensors', { headers });
        if (response.ok) {
          const data = await response.json();
          if (data.devices) {
            const newMap: { [key: number]: { [address: number]: number } } = {};
            Object.keys(data.devices).forEach(devIdStr => {
              const devId = Number(devIdStr);
              const addrMap = data.devices[devIdStr];
              newMap[devId] = {};
              Object.keys(addrMap).forEach(addrStr => {
                newMap[devId][Number(addrStr)] = addrMap[addrStr];
              });
            });
            setSensorsMap(newMap);
          }
        }
      } catch (error) {
        setLogMessage('Uyarı: Sensör verisi alınamadı.');
      }
    };

    fetchSensors();
    const interval = setInterval(fetchSensors, 2000);
    return () => clearInterval(interval);
  }, [token]);

  // Röle Kontrol İstek Atma
  const handleRelayToggle = async (deviceId: number, relayId: number, newState: boolean) => {
    if (!token) return;
    try {
      const response = await fetch('http://localhost:8080/api/relay', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ device_id: deviceId, id: relayId, state: newState })
      });

      if (response.ok) {
        setRelayStates(prev => ({ ...prev, [`${deviceId}-${relayId}`]: newState }));
        setLogMessage(`Cihaz #${deviceId} Röle #${relayId} başarıyla ${newState ? 'AÇILDI (1)' : 'KAPATILDI (0)'} olarak ayarlandı.`);
      } else if (response.status === 403) {
        setLogMessage(`Yetki Hatası (403): Bu röleyi değiştirmek için izniniz yok!`);
      } else {
        setLogMessage(`Hata: Komut reddedildi.`);
      }
    } catch (error) {
      setLogMessage(`Hata: İstek iletilemedi.`);
    }
  };

  const handleLoginSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoginError('');
    const res = await login(usernameInput, passwordInput);
    if (!res.success) {
      setLoginError(res.message || 'Giriş başarısız');
    }
  };

  if (!token || !user) {
    return (
      <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center p-4">
        <div className="bg-gray-800 p-8 rounded-lg shadow-xl border border-gray-700 w-full max-w-md">
          <h2 className="text-2xl font-bold mb-6 text-center text-cyan-400">SCADA Yetkili Giriş</h2>
          {loginError && <div className="bg-red-900/50 border border-red-500 text-red-200 p-3 rounded mb-4 text-sm">{loginError}</div>}
          <form onSubmit={handleLoginSubmit} className="space-y-4">
            <div>
              <label className="block text-sm text-gray-400 mb-1">Kullanıcı Adı</label>
              <input type="text" value={usernameInput} onChange={(e) => setUsernameInput(e.target.value)} className="w-full bg-gray-900 border border-gray-700 rounded p-2.5 text-white focus:outline-none focus:border-cyan-500" required />
            </div>
            <div>
              <label className="block text-sm text-gray-400 mb-1">Şifre</label>
              <input type="password" value={passwordInput} onChange={(e) => setPasswordInput(e.target.value)} className="w-full bg-gray-900 border border-gray-700 rounded p-2.5 text-white focus:outline-none focus:border-cyan-500" required />
            </div>
            <button type="submit" className="w-full bg-cyan-600 hover:bg-cyan-500 text-white font-medium py-2.5 rounded transition shadow-lg">Giriş Yap</button>
          </form>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-900 text-white p-8">
      <div className="max-w-4xl mx-auto space-y-6">
        <header className="border-b border-gray-700 pb-4 flex justify-between items-center">
          <div>
            <h1 className="text-2xl font-bold">Donanım İzleme ve Kontrol Paneli</h1>
            <p className="text-sm text-gray-400">
              Hoş geldin, <span className="text-cyan-400 font-semibold">{user.username}</span> ({user.isSuperuser ? 'Sistem Yöneticisi' : 'Operatör'})
            </p>
          </div>
          <button onClick={logout} className="bg-red-600/80 hover:bg-red-600 text-white px-4 py-2 rounded text-sm font-medium transition">Çıkış Yap</button>
        </header>

        {/* Cihaz Bazlı Bölmeler */}
        <div className="space-y-6">
          {devices.map((device) => {
            const devSensorsValues = sensorsMap[device.id] || {};

            return (
              <div key={device.id} className="bg-gray-800 p-6 rounded-lg shadow-lg border border-gray-700 space-y-5">
                {/* Cihaz Başlığı */}
                <div className="border-b border-gray-700 pb-3 flex justify-between items-center">
                  <h2 className="text-lg font-bold text-cyan-400 flex items-center gap-2">
                    🔌 {device.name} <span className="text-xs text-gray-400 font-normal">({device.url})</span>
                  </h2>
                </div>

                {/* 1. Dinamik Sensör Kutuları*/}
                <div className={`grid grid-cols-${Math.min(device.sensors.length, 2)} gap-4`}>
                  {device.sensors.map((sensor) => {
                    const val = devSensorsValues[sensor.address] !== undefined ? devSensorsValues[sensor.address] : 0;
                    return (
                      <div key={sensor.address} className="bg-gray-900 p-4 rounded-lg border border-gray-700">
                        <span className="text-xs text-gray-400">{sensor.name}</span>
                        <p className="text-2xl font-semibold mt-1 text-cyan-400">
                          {val.toFixed(2)} {sensor.unit}
                        </p>
                      </div>
                    );
                  })}
                </div>

                {/* 2. Bu Cihazın Röle Kontrolleri */}
                <div>
                  <h3 className="text-sm font-medium text-gray-300 mb-3">Cihaz Röle Kontrolleri</h3>
                  {device.relays.length === 0 ? (
                    <p className="text-xs text-gray-500 italic">Bu cihaza ait tanımlı röle bulunmuyor.</p>
                  ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      {device.relays.map((relay) => {
                        const stateKey = `${device.id}-${relay.id}`;
                        const isOn = relayStates[stateKey] || false;
                        const permissionKey = `accounts.device${device.id}_relay${relay.id}_change`;
                        const canControl = hasPermission(permissionKey);

                        return (
                          <div key={relay.id} className="flex items-center justify-between bg-gray-700/50 p-4 rounded-lg border border-gray-600">
                            <div>
                              <span className="font-medium">Röle #{relay.id}</span>
                              <p className="text-xs text-gray-400">{relay.name} (Adres: {relay.coil_address})</p>
                              {!canControl && (
                                <span className="text-xs text-red-400 block mt-1">⚠️ Bu röle için yetkiniz yok</span>
                              )}
                            </div>
                            <button 
                              onClick={() => handleRelayToggle(device.id, relay.id, !isOn)} 
                              disabled={!canControl}
                              className={`px-4 py-2 rounded font-medium transition shadow-md ${
                                !canControl 
                                  ? 'bg-gray-600 text-gray-400 cursor-not-allowed opacity-50' 
                                  : isOn 
                                  ? 'bg-red-600 hover:bg-red-500 text-white' 
                                  : 'bg-green-600 hover:bg-green-500 text-white'
                              }`}
                            >
                              {isOn ? 'Kapat (0)' : 'Aç (1)'}
                            </button>
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>

        {/* Log Paneli */}
        <div className="bg-black p-4 rounded-lg font-mono text-sm text-green-400 border border-gray-800 shadow-inner">
          <p className="text-gray-500">// Sistem Günlükleri & Olay Akışı</p>
          <p className="mt-1">{logMessage}</p>
        </div>
      </div>
    </div>
  );
}

export default App;