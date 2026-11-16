# Endüstriyel Donanım İzleme ve Kontrol Platformu (SCADA / Modbus TCP Gateway)

Çoklu PLC cihazlarıyla **Modbus TCP** protokolü üzerinden haberleşen, gerçek zamanlı sensör verilerini okuyan, read-back doğrulamalı röle kontrolü sağlayan ve mikroservis mimarisiyle tasarlanmış endüstriyel izleme ve kontrol platformudur.

---

## 1. Kurulum ve Çalıştırma (Setup)

### Ön Gereksinimler
* Docker & Docker Compose

### Ortam Değişkenleri (`.env`)
Proje ana dizininde bir `.env` dosyası oluşturun ve JWT doğrulama anahtarını ile veritabanı bilgilerini tanımlayın:

### Tek Komutla Başlatma

Tüm servisleri (Go Backend, Django Auth, React Frontend, Modbus Simülatörleri, Prometheus ve Grafana) derleyip ayağa kaldırmak için:

```bash
docker compose up -d --build

```

### Servis Erişim Adresleri

| Servis | Adres | Açıklama |
| --- | --- | --- |
| **Frontend Arayüzü** | `http://localhost:3000` | Canlı sensör izleme ve röle kontrol paneli |
| **Go Backend API** | `http://localhost:8080` | REST API Gateway ve `/metrics` uç noktası |
| **Django Auth API** | `http://localhost:8000` | JWT token üretimi ve yönetim paneli (`/admin`) |
| **Prometheus** | `http://localhost:9090` | Metrik sorgulama ve hedef (target) durumu |
| **Grafana** | `http://localhost:3001` | Görsel izleme panelleri (Varsayılan: `admin` / `admin`) |

---

## 2. Mimari Özet (Architecture Summary)

Platform, izole bir Docker ağı (`modbus-net`) üzerinde çalışan birbirine entegre servislerden oluşur:

* **Go API Gateway & Poller (`backend`):** Eşzamanlı (*concurrent*) goroutine mimarisi ile birden fazla PLC cihazından periyodik olarak sensör verilerini okur, bağlantı kopmalarında otomatik yeniden bağlanma (*reconnect with retry*) sağlar, Modbus Coil yazma/geri okuma (*read-back*) işlemlerini yönetir ve Prometheus metriklerini dışa aktarır.
* **Kimlik Doğrulama & Yetkilendirme Servisi (`django-auth`):** Django ve SimpleJWT altyapısı ile kullanıcı oturumlarını yönetir ve özel tanımlı izinleri (*permissions*) içeren JWT token üretir.
* **Web Kontrol Paneli (`frontend`):** React tabanlı arayüz üzerinden canlı sensör verilerini sunar ve kullanıcının sahip olduğu yetkilere göre röle kontrol bileşenlerini yönetir.
* **Modbus TCP Simülatörü (`simulator`):** Saha PLC donanımlarını (Kazan Dairesi PLC-1, Soğuk Hava Deposu PLC-2) simüle eden sanal sunucular.
* **Gözlemlenebilirlik Yığını (`prometheus` & `grafana`):** Modbus bağlantı durumlarını, okuma gecikme sürelerini, röle tetiklenme sayaçlarını ve HTTP istek trafiğini gerçek zamanlı olarak izler.

*(Not: Cihaz, sensör ve röle adres haritaları proje dizinindeki `configs/config.json.example` dosyasında tanımlıdır.)*

---

## 3. Yetki Kataloğu (Capability Catalog)

Sistemde cihaz ve röle bazlı granüler yetkilendirme uygulanmaktadır. Django tarafında tanımlanan ve JWT token içerisine gömülen 4 temel yetki şunlardır:

| Yetki Kodu (Permission) | Açıklama | Etki Alanı |
| --- | --- | --- |
| `device1_relay1_change` | Cihaz 1 - Röle 1 Durumunu Değiştirebilir | Kazan Dairesi PLC-1 (Ana Isıtıcı) |
| `device1_relay2_change` | Cihaz 1 - Röle 2 Durumunu Değiştirebilir | Kazan Dairesi PLC-1 (Sirkülasyon Pompası) |
| `device2_relay1_change` | Cihaz 2 - Röle 1 Durumunu Değiştirebilir | Soğuk Hava Deposu PLC-2 (Kompresör) |
| `device2_relay2_change` | Cihaz 2 - Röle 2 Durumunu Değiştirebilir | Soğuk Hava Deposu PLC-2 (Fan Grubu 1) |

---

## 4. Bilinen Kısıtlamalar, Tamamlanmamış Kısımlar ve Gerekçeleri

Projenin mevcut sürümünde zaman kısıtları ve mimari öncelikler nedeniyle kapsam dışında bırakılan veya geliştirilmesi planlanan alanlar aşağıda açıklanmıştır:

* **Statik Konfigürasyon ve Veritabanı Eksikliği:** Cihaz, sensör ve röle tanımları şu an merkezi bir ilişkisel veritabanı yerine statik `configs/config.json` dosyasından okunmaktadır. Temel Modbus haberleşme ve kontrol akışını önceliklendirmek amacıyla genel bir veritabanı katmanı kurulmamıştır.
* **Sınırlı İşlem ve Denetim Kaydı (Audit Logging):** Yapılan röle tetikleme ve sistem işlemleri kalıcı bir veritabanına kaydedilmemektedir. Temel bir loglama fonksiyonu (`pkg.LogAction`) oluşturulup birkaç noktaya entegre edilmiş ancak tüm uç noktalara ve kalıcı depolamaya yaygınlaştırılması zaman kısıtı nedeniyle tamamlanamamıştır.
* **İstemci Taraflı (Frontend) Yetki Kısıtlaması ve Token İşleme:** Django tarafından üretilen ve Go Backend üzerinden React arayüzüne iletilen JWT token, şu an frontend tarafında çözümlenerek (*decode*) kullanıcı yetkilerine göre arayüz bileşenlerini kısıtlamaktadır. Güvenlik derinliği (*defense-in-depth*) açısından bu kısıtlamaların doğrudan Go Backend middleware katmanında zorunlu kılınması ve token aktarım/doğrulama akışının iyileştirilmesi gerekmektedir.
* **Statik Yetki Tanımlamaları:** Yeni yetki tanımlamaları Django model.py sınıfı içerisinde kod seviyesinde (`hardcoded`) tanımlanmıştır. Yeni cihaz veya röle eklendiğinde kod değişikliği gerektirmeyen, dinamik ve daha güvenli bir veritabanı tabanlı yetkilendirme yapısına geçilmesi planlanmaktadır.
* Sıcaklık tabanlı otomatik fan kontrolü (*histerezis* döngüsü ve çalışma zamanı eşik yönetimi)
* Çok kanallı termokupl (*thermocouple*) modülü desteği
* Analog / PWM çıkışı aracılığıyla değişken fan hızı kontrolü


## 5. Test Kullanıcı Bilgileri (Test Login Credentials)

Sistem ayağa kalktığında aşağıdaki test hesapları ve yetki kombinasyonları otomatik olarak oluşturulur:

| Rol / Senaryo | Kullanıcı Adı | Şifre | Sahip Olduğu Yetkiler (Permissions) |
| :--- | :--- | :--- | :--- |
| **Tam Yetkili (Admin)** | `admin` | `Admin123!` | Tüm Cihaz ve Röle Yetkileri (`device1_relay1_change`, `device1_relay2_change`, `device2_relay1_change`, `device2_relay2_change`) |
| **Kazan Dairesi Operatörü** | `operator_kazan` | `Operator123!` | Yalnızca Cihaz 1 Röleleri (`device1_relay1_change`, `device1_relay2_change`) |
| **İzleyici (Salt Okunur)** | `viewer` | `Viewer123!` | Yetki yok (Yalnızca sensör verilerini izleyebilir, röle tetikleyemez) |


```

```