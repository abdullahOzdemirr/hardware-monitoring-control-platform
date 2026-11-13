import React, { useState, useEffect } from 'react';
import { SensorData } from './types';

interface RelayConfig {
  id: number;
  name: string;
  coil_address: number;
}

export function App() {
  const [sensors, setSensors] = useState<SensorData>({
    temperature: 0,
    humidity: 0,
    timestamp: '-'
  });
  
  const [relays, setRelays] = useState<RelayConfig[]>([]);
  const [relayStates, setRelayStates] = useState<{ [key: number]: boolean }>({});
  const [logMessage, setLogMessage] = useState<string>('Sistem başlatıldı. Backend bekleniyor...');

  // Backend'den Verileri Çekme
  useEffect(() => {
    // 1. Röle listesini config'den al
    fetch('http://localhost:8080/api/relays')
      .then(res => res.json())
      .then((data: RelayConfig[]) => {
        setRelays(data);
        // Başlangıç durumları false
        const initialStates: { [key: number]: boolean } = {};
        data.forEach(r => { initialStates[r.id] = false; });
        setRelayStates(initialStates);
      })
      .catch(() => setLogMessage('Uyarı: Röle konfigürasyonu alınamadı.'));

    // 2. Periyodik Sensör Polling
    const fetchData = async () => {
      try {
        const response = await fetch('http://localhost:8080/api/sensors');
        if (response.ok) {
          const data = await response.json();
          setSensors({
            temperature: data.temperature || 0,
            humidity: data.humidity || 0,
            timestamp: new Date().toLocaleTimeString()
          });
        }
      } catch (error) {
        setLogMessage('Uyarı: Go backend bağlantısı kurulamadı.');
      }
    };

    fetchData();
    const interval = setInterval(fetchData, 2000);
    return () => clearInterval(interval);
  }, []);

  const handleRelayToggle = async (relayId: number, newState: boolean) => {
    try {
      const response = await fetch('http://localhost:8080/api/relay', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: relayId, state: newState })
      });

      if (response.ok) {
        setRelayStates(prev => ({ ...prev, [relayId]: newState }));
        setLogMessage(`Röle #${relayId} başarıyla ${newState ? 'AÇILDI (1)' : 'KAPATILDI (0)'} olarak ayarlandı.`);
      } else {
        setLogMessage(`Hata: Röle #${relayId} komutu reddedildi.`);
      }
    } catch (error) {
      setLogMessage(`Hata: Röle #${relayId} isteği iletilemedi.`);
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 text-white p-8">
      <div className="max-w-4xl mx-auto space-y-6">
        <header className="border-b border-gray-700 pb-4">
          <h1 className="text-2xl font-bold">Donanım İzleme ve Kontrol Paneli</h1>
          <p className="text-sm text-gray-400">React 18 + TypeScript + Go Backend Entegrasyonu</p>
        </header>

        {/* Sensör Kartları */}
        <div className="grid grid-cols-2 gap-4">
          <div className="bg-gray-800 p-6 rounded-lg shadow border border-gray-700">
            <h2 className="text-sm text-gray-400">Sıcaklık</h2>
            <p className="text-3xl font-semibold mt-2 text-cyan-400">{sensors.temperature.toFixed(2)} °C</p>
          </div>
          <div className="bg-gray-800 p-6 rounded-lg shadow border border-gray-700">
            <h2 className="text-sm text-gray-400">Nem</h2>
            <p className="text-3xl font-semibold mt-2 text-cyan-400">{sensors.humidity.toFixed(2)} %</p>
          </div>
        </div>

        {/* Dinamik Röle Kontrol Paneli */}
        <div className="bg-gray-800 p-6 rounded-lg shadow border border-gray-700">
          <h2 className="text-lg font-medium mb-4">Röle Kontrolleri (Config'den Dinamik Yüklenir)</h2>
          <div className="grid grid-cols-2 gap-4">
            {relays.map((relay) => {
              const isOn = relayStates[relay.id] || false;
              return (
                <div key={relay.id} className="flex items-center justify-between bg-gray-700 p-4 rounded border border-gray-600">
                  <div>
                    <span className="font-medium">Röle #{relay.id}</span>
                    <p className="text-xs text-gray-400">{relay.name} (Adres: {relay.coil_address})</p>
                  </div>
                  <button
                    onClick={() => handleRelayToggle(relay.id, !isOn)}
                    className={`px-4 py-2 rounded font-medium transition shadow-md ${
                      isOn 
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
        </div>

        {/* Durum / Log Paneli */}
        <div className="bg-black p-4 rounded-lg font-mono text-sm text-green-400 border border-gray-800 shadow-inner">
          <p className="text-gray-500">// Sistem Günlükleri & Olay Akışı</p>
          <p className="mt-1">{logMessage}</p>
        </div>
      </div>
    </div>
  );
}

export default App;