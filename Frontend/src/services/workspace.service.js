import axiosInstance from '../api/axiosInstance';

const workspaceService = {
  getAll: () => axiosInstance.get('/organizations'),

  getBySlug: (slug) => axiosInstance.get(`/organizations/${slug}`),

  create: (data) => axiosInstance.post('/organizations', data),

  update: (slug, data) => axiosInstance.put(`/organizations/${slug}`, data),

  delete: (slug) => axiosInstance.delete(`/organizations/${slug}`),

  getMembers: (slug) => axiosInstance.get(`/organizations/${slug}/members`),

  getBoardMembers: (slug) => axiosInstance.get(`/organizations/${slug}/board-members`),

  inviteMember: (slug, email, role = 'member') =>
    axiosInstance.post(`/organizations/${slug}/members`, { email, role }),

  updateMemberRole: (slug, userId, role) =>
    axiosInstance.put(`/organizations/${slug}/members/${userId}`, { role }),

  removeMember: (slug, userId) =>
    axiosInstance.delete(`/organizations/${slug}/members/${userId}`),

  transferOwnership: (slug, userId) =>
    axiosInstance.post(`/organizations/${slug}/transfer-ownership`, { user_id: userId }),

  updateBoardMemberRole: (slug, boardId, userId, role) =>
    axiosInstance.put(`/organizations/${slug}/boards/${boardId}/members/${userId}`, { role }),
};

export default workspaceService;
