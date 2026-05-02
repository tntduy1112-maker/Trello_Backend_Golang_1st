import axiosInstance from '../api/axiosInstance';

const commentService = {
  list: (cardId) => axiosInstance.get(`/cards/${cardId}/comments`),

  create: (cardId, content, parentId = null) =>
    axiosInstance.post(`/cards/${cardId}/comments`, { content, parent_id: parentId }),

  update: (commentId, content) => axiosInstance.put(`/comments/${commentId}`, { content }),

  delete: (commentId) => axiosInstance.delete(`/comments/${commentId}`),
};

export default commentService;
