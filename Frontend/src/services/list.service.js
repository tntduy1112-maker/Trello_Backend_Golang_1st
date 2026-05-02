import axiosInstance from '../api/axiosInstance';

const listService = {
  create: (boardId, title) => axiosInstance.post(`/boards/${boardId}/lists`, { title }),

  update: (listId, data) => axiosInstance.put(`/lists/${listId}`, data),

  archive: (listId) => axiosInstance.post(`/lists/${listId}/archive`),

  restore: (listId) => axiosInstance.post(`/lists/${listId}/restore`),

  move: (listId, position) => axiosInstance.post(`/lists/${listId}/move`, { position }),

  copy: (listId, title) => axiosInstance.post(`/lists/${listId}/copy`, { title }),
};

export default listService;
