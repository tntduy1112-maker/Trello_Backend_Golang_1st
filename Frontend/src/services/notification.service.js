import axiosInstance from '../api/axiosInstance';

const notificationService = {
  getAll: (page = 1, limit = 20) =>
    axiosInstance.get('/notifications', { params: { page, limit } }),

  getUnreadCount: () => axiosInstance.get('/notifications/unread-count'),

  markAsRead: (notificationId) => axiosInstance.post(`/notifications/${notificationId}/read`),

  markAllAsRead: () => axiosInstance.post('/notifications/read-all'),

  delete: (notificationId) => axiosInstance.delete(`/notifications/${notificationId}`),
};

export default notificationService;
