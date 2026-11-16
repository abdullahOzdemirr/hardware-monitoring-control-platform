from django.db import models

class ScadaDevicePermissions(models.Model):
    """
    SCADA cihaz yetkilerini Django admin panelinde görebilmek 
    ve kullanıcılara/gruplara atayabilmek için sanal model.
    """
    class Meta:
        managed = False
        permissions = [
            ("device1_relay1_change", "Cihaz 1 - Röle 1 Durumunu Değiştirebilir"),
            ("device1_relay2_change", "Cihaz 1 - Röle 2 Durumunu Değiştirebilir"),
            ("device2_relay1_change", "Cihaz 2 - Röle 1 Durumunu Değiştirebilir"),
            ("device2_relay2_change", "Cihaz 2 - Röle 2 Durumunu Değiştirebilir"),
        ]