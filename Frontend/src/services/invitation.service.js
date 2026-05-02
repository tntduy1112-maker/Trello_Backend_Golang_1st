import axiosInstance from '../api/axiosInstance';
import axios from 'axios';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

const invitationService = {
  // Public endpoints - no auth required
  getByToken: (token) => axios.get(`${API_URL}/invitations/${token}`),

  acceptWithPassword: (token, data) =>
    axios.post(`${API_URL}/invitations/${token}/accept-with-password`, data),

  // Protected endpoints
  accept: (token) => axiosInstance.post(`/invitations/${token}/accept`),

  decline: (token) => axiosInstance.post(`/invitations/${token}/decline`),
};

export default invitationService;
