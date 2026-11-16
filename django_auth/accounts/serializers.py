from rest_framework_simplejwt.serializers import TokenObtainPairSerializer

class CustomTokenObtainPairSerializer(TokenObtainPairSerializer):
    @classmethod
    def get_token(cls, user):
        token = super().get_token(user)
        
        # Kullanıcının sahip olduğu tüm izinleri alıyoruz
        permissions = list(user.get_all_permissions())
        
        # Token payload'ına ekliyoruz ki Go ve React okuyabilsin
        token['username'] = user.username
        token['permissions'] = permissions
        token['is_superuser'] = user.is_superuser
        
        return token