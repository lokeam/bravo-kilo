import { useMutation, useQuery } from '@tanstack/react-query';
import axios from 'axios';

interface LoginResponse {
  token: string;
}

interface LoginData {
  username: string;
  password: string;
}

const login = async (loginData: LoginData): Promise<LoginResponse> => {
  const response = await axios.post('google-signin', loginData);

  document.cookie = `token=${response.data.token}; HttpOnly; Secure`
  return response.data;
}

export const useLogin = () => {
  return useMutation<LoginResponse, Error, LoginData>(login);
}

const fetchProtectedData = async () => {
  const token = document.cookie.split('; ').find(row => row.startsWith('token='))?.split('=')[1];
  const response = await axios.get('/api/protected', {
    headers: { Authorization: `Bearer ${token}` },
  });

  return response.data;
};

export const useProtectedData = () => {
  return useQuery('protectedData', fetchProtectedData);
}
