import axiosInstance from '../api/axiosInstance';

const boardService = {
  getByWorkspace: (workspaceSlug) => axiosInstance.get(`/organizations/${workspaceSlug}/boards`),

  getById: (boardId) => axiosInstance.get(`/boards/${boardId}`),

  create: (workspaceSlug, data) => axiosInstance.post(`/organizations/${workspaceSlug}/boards`, data),

  update: (boardId, data) => axiosInstance.put(`/boards/${boardId}`, data),

  delete: (boardId) => axiosInstance.delete(`/boards/${boardId}`),

  close: (boardId) => axiosInstance.post(`/boards/${boardId}/close`),

  reopen: (boardId) => axiosInstance.post(`/boards/${boardId}/reopen`),

  getMembers: (boardId) => axiosInstance.get(`/boards/${boardId}/members`),

  inviteMember: (boardId, email, role = 'member') =>
    axiosInstance.post(`/boards/${boardId}/members`, { email, role }),

  getInvitations: (boardId) => axiosInstance.get(`/boards/${boardId}/invitations`),

  revokeInvitation: (boardId, invitationId) =>
    axiosInstance.delete(`/boards/${boardId}/invitations/${invitationId}`),

  getLabels: (boardId) => axiosInstance.get(`/boards/${boardId}/labels`),

  createLabel: (boardId, data) => axiosInstance.post(`/boards/${boardId}/labels`, data),

  updateLabel: (labelId, data) => axiosInstance.put(`/labels/${labelId}`, data),

  deleteLabel: (labelId) => axiosInstance.delete(`/labels/${labelId}`),

  getActivity: (boardId, page = 1, limit = 20) =>
    axiosInstance.get(`/boards/${boardId}/activity`, { params: { page, limit } }),
};

export default boardService;
