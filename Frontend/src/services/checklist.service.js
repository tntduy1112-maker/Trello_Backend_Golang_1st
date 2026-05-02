import axiosInstance from '../api/axiosInstance';

const checklistService = {
  list: (cardId) => axiosInstance.get(`/cards/${cardId}/checklists`),

  create: (cardId, title) => axiosInstance.post(`/cards/${cardId}/checklists`, { title }),

  update: (checklistId, title) => axiosInstance.put(`/checklists/${checklistId}`, { title }),

  delete: (checklistId) => axiosInstance.delete(`/checklists/${checklistId}`),

  createItem: (checklistId, title) =>
    axiosInstance.post(`/checklists/${checklistId}/items`, { title }),

  updateItem: (itemId, data) => axiosInstance.put(`/checklist-items/${itemId}`, data),

  deleteItem: (itemId) => axiosInstance.delete(`/checklist-items/${itemId}`),

  toggleItem: (itemId) => axiosInstance.post(`/checklist-items/${itemId}/toggle`),
};

export default checklistService;
