import Cookies from 'js-cookie';

export const getToken = (): string | null => {
  const token = Cookies.get('token');
  console.log('Token from cookies in getToken:', token);
  return token || null;
};
