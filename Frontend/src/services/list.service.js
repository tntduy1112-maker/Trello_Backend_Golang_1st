import axiosInstance from '../api/axiosInstance';

const listService = {
  create: (boardId, title) => axiosInstance.post(`/boards/${boardId}/lists`, { title }),

  update: (listId, data) => axiosInstance.put(`/lists/${listId}`, data),

  move: (listId, position) => axiosInstance.put(`/lists/${listId}/move`, { position }),

  archive: (listId) => axiosInstance.delete(`/lists/${listId}`),
};

export default listService;
