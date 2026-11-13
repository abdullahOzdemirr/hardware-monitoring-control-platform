import { createContext, useContext, useState } from 'react';

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
  // Şimdilik tam yetkili admin var
  // İleride kullanıcı oluştururken etiketler buraya dinamik olarak atanacak.
  const [user, setUser] = useState({
    username: 'admin',
    role: 'Admin',
    permissions: ['read:sensors', 'write:relays', 'manage:system']
  });

  const hasPermission = (permission) => {
    return user.permissions.includes(permission);
  };

  return (
    <AuthContext.Provider value={{ user, hasPermission }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => useContext(AuthContext);