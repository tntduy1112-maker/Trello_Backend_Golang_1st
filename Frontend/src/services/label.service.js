import axiosInstance from '../api/axiosInstance';

const labelService = {
  getByBoard: (boardId) => axiosInstance.get(`/boards/${boardId}/labels`),

  create: (boardId, data) => axiosInstance.post(`/boards/${boardId}/labels`, data),

  update: (labelId, data) => axiosInstance.put(`/labels/${labelId}`, data),

  delete: (labelId) => axiosInstance.delete(`/labels/${labelId}`),
};

export default labelService;
