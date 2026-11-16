from django.core.management.base import BaseCommand
from django.contrib.auth.models import User, Permission

class Command(BaseCommand):
    help = 'Test kullanıcılarını ve özel SCADA yetkilerini otomatik oluşturur'

    def handle(self, *args, **kwargs):
        test_accounts = [
            {
                "username": "admin",
                "password": "Admin123!",
                "is_superuser": True,
                "is_staff": True,
                "perms": [
                    "device1_relay1_change",
                    "device1_relay2_change",
                    "device2_relay1_change",
                    "device2_relay2_change",
                ],
            },
            {
                "username": "operator_kazan",
                "password": "Operator123!",
                "is_superuser": False,
                "is_staff": False,
                "perms": [
                    "device1_relay1_change",
                    "device1_relay2_change",
                ],
            },
            {
                "username": "viewer",
                "password": "Viewer123!",
                "is_superuser": False,
                "is_staff": False,
                "perms": [],
            },
        ]

        for acc in test_accounts:
            user, created = User.objects.get_or_create(username=acc["username"])
            user.set_password(acc["password"])
            user.is_superuser = acc["is_superuser"]
            user.is_staff = acc["is_staff"]
            user.save()

            perms = Permission.objects.filter(codename__in=acc["perms"])
            user.user_permissions.set(perms)

            status = "oluşturuldu" if created else "güncellendi"
            self.stdout.write(self.style.SUCCESS(f"Kullanıcı '{acc['username']}' {status}."))