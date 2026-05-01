import axiosInstance from '../api/axiosInstance';

const cardService = {
  create: (listId, title) => axiosInstance.post(`/lists/${listId}/cards`, { title }),

  getById: (cardId) => axiosInstance.get(`/cards/${cardId}`),

  update: (cardId, data) => axiosInstance.put(`/cards/${cardId}`, data),

  move: (cardId, listId, position) =>
    axiosInstance.put(`/cards/${cardId}/move`, { list_id: listId, position }),

  archive: (cardId) => axiosInstance.delete(`/cards/${cardId}`),

  assign: (cardId, userId) => axiosInstance.post(`/cards/${cardId}/assign`, { user_id: userId }),

  unassign: (cardId) => axiosInstance.delete(`/cards/${cardId}/assign`),

  toggleComplete: (cardId) => axiosInstance.post(`/cards/${cardId}/complete`),

  addLabel: (cardId, labelId) => axiosInstance.post(`/cards/${cardId}/labels/${labelId}`),

  removeLabel: (cardId, labelId) => axiosInstance.delete(`/cards/${cardId}/labels/${labelId}`),

  getComments: (cardId) => axiosInstance.get(`/cards/${cardId}/comments`),

  addComment: (cardId, content) => axiosInstance.post(`/cards/${cardId}/comments`, { content }),

  updateComment: (commentId, content) => axiosInstance.put(`/comments/${commentId}`, { content }),

  deleteComment: (commentId) => axiosInstance.delete(`/comments/${commentId}`),

  getChecklists: (cardId) => axiosInstance.get(`/cards/${cardId}/checklists`),

  createChecklist: (cardId, title) => axiosInstance.post(`/cards/${cardId}/checklists`, { title }),

  deleteChecklist: (checklistId) => axiosInstance.delete(`/checklists/${checklistId}`),

  createChecklistItem: (checklistId, title) =>
    axiosInstance.post(`/checklists/${checklistId}/items`, { title }),

  toggleChecklistItem: (itemId) => axiosInstance.post(`/checklist-items/${itemId}/toggle`),

  deleteChecklistItem: (itemId) => axiosInstance.delete(`/checklist-items/${itemId}`),

  getAttachments: (cardId) => axiosInstance.get(`/cards/${cardId}/attachments`),

  uploadAttachment: (cardId, file) => {
    const formData = new FormData();
    formData.append('file', file);
    return axiosInstance.post(`/cards/${cardId}/attachments`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  },

  deleteAttachment: (attachmentId) => axiosInstance.delete(`/attachments/${attachmentId}`),

  setCover: (attachmentId) => axiosInstance.post(`/attachments/${attachmentId}/cover`),

  removeCover: (cardId) => axiosInstance.delete(`/cards/${cardId}/cover`),

  getActivity: (cardId) => axiosInstance.get(`/cards/${cardId}/activity`),
};

export default cardService;
