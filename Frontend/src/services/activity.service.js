import axiosInstance from '../api/axiosInstance';

const activityService = {
  listByCard: (cardId, page = 1, limit = 20) =>
    axiosInstance.get(`/cards/${cardId}/activity`, { params: { page, limit } }),

  listByBoard: (boardId, page = 1, limit = 20) =>
    axiosInstance.get(`/boards/${boardId}/activity`, { params: { page, limit } }),
};

export default activityService;
