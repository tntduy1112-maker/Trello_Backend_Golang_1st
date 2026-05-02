import axiosInstance from '../api/axiosInstance';

const attachmentService = {
  list: (cardId) => axiosInstance.get(`/cards/${cardId}/attachments`),

  upload: (cardId, file) => {
    const formData = new FormData();
    formData.append('file', file);
    return axiosInstance.post(`/cards/${cardId}/attachments`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  },

  delete: (attachmentId) => axiosInstance.delete(`/attachments/${attachmentId}`),

  setCover: (attachmentId) => axiosInstance.post(`/attachments/${attachmentId}/cover`),

  removeCover: (cardId) => axiosInstance.delete(`/cards/${cardId}/cover`),

  getDownloadUrl: (attachmentId) => axiosInstance.get(`/attachments/${attachmentId}/download`),
};

export default attachmentService;
