from rest_framework_simplejwt.serializers import TokenObtainPairSerializer

class CustomTokenObtainPairSerializer(TokenObtainPairSerializer):
    @classmethod
    def get_token(cls, user):
        token = super().get_token(user)
        
        permissions = list(user.get_all_permissions())
        
        token['username'] = user.username
        token['permissions'] = permissions
        token['is_superuser'] = user.is_superuser
        
        return token