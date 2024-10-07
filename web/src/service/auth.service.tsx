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
  try {
    const response = await axios.post('google/signin', loginData);
    document.cookie = `token=${response.data.token}; HttpOnly; Secure`;
    console.log('Login successful, token set:', response.data.token);
    return response.data;
  } catch (error) {
    console.error('Login error:', error);
    throw error;
  }
}

export const useLogin = () => {
  return useMutation<LoginResponse, Error, LoginData>({
    mutationFn: login,
  });
}

const fetchProtectedData = async () => {
  try {
    const token = document.cookie.split('; ').find(row => row.startsWith('token='))?.split('=')[1];
    const response = await axios.get('/api/protected', {
      headers: { Authorization: `Bearer ${token}` },
    });
    console.log('Protected data fetched:', response.data);
    return response.data;
  } catch (error) {
    console.error('Error fetching protected data:', error);
    throw error;
  }
};

export const useProtectedData = () => {
  return useQuery({
    queryKey: ['protectedData'],
    queryFn: fetchProtectedData,
  });
};
