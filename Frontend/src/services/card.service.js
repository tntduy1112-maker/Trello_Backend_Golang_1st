import axiosInstance from '../api/axiosInstance';

const cardService = {
  create: (listId, title) => axiosInstance.post(`/lists/${listId}/cards`, { title }),

  getById: (cardId) => axiosInstance.get(`/cards/${cardId}`).then((res) => ({ data: res.data.data })),

  update: (cardId, data) => axiosInstance.put(`/cards/${cardId}`, data),

  archive: (cardId) => axiosInstance.post(`/cards/${cardId}/archive`),

  restore: (cardId) => axiosInstance.post(`/cards/${cardId}/restore`),

  move: (cardId, listId, position) =>
    axiosInstance.post(`/cards/${cardId}/move`, { list_id: listId, position }),

  assign: (cardId, userId) => axiosInstance.post(`/cards/${cardId}/assign`, { user_id: userId }),

  unassign: (cardId) => axiosInstance.post(`/cards/${cardId}/unassign`),

  markComplete: (cardId) => axiosInstance.post(`/cards/${cardId}/complete`),

  markIncomplete: (cardId) => axiosInstance.post(`/cards/${cardId}/incomplete`),

  addLabel: (cardId, labelId) => axiosInstance.post(`/cards/${cardId}/labels/${labelId}`),

  removeLabel: (cardId, labelId) => axiosInstance.delete(`/cards/${cardId}/labels/${labelId}`),
};

export default cardService;
