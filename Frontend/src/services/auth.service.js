import axiosInstance from '../api/axiosInstance';

const authService = {
  register: (data) => axiosInstance.post('/auth/register', data),

  verifyEmail: (email, otp) => axiosInstance.post('/auth/verify-email', { email, otp }),

  resendVerification: (email) => axiosInstance.post('/auth/resend-verification', { email }),

  login: (credentials) => axiosInstance.post('/auth/login', credentials),

  logout: () => axiosInstance.post('/auth/logout'),

  logoutAll: () => axiosInstance.post('/auth/logout-all'),

  forgotPassword: (email) => axiosInstance.post('/auth/forgot-password', { email }),

  resetPassword: (token, password) => axiosInstance.post('/auth/reset-password', { token, password }),

  getCurrentUser: () => axiosInstance.get('/auth/me'),

  updateProfile: (data) => axiosInstance.put('/auth/me', data),

  updateAvatar: (file) => {
    const formData = new FormData();
    formData.append('avatar', file);
    return axiosInstance.put('/auth/me', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  },
};

export default authService;
